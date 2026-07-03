package main

import _ "github.com/lib/pq"

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/CarolineMillan/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type apiConfig struct {
	// atomic int lets us safely access and update data over multiple goroutines
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// increment the fileserverHits counter
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})

}

func (cfg *apiConfig) printHitsHandler(res http.ResponseWriter, req *http.Request) {
	// prints the number of fileserverHits
	num := cfg.fileserverHits.Load()
	txt := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", num)

	res.Header().Add("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(txt))
}

func (cfg *apiConfig) resetHitsHandler(res http.ResponseWriter, req *http.Request) {
	// resets the number of fileserverHits and users in the database
	txt := strconv.AppendInt([]byte("Hits: "), 0, 10)

	// check we're in the dev environment
	if cfg.platform == "dev" {
		cfg.db.ResetUsers(req.Context())
	} else {
		txt2 := "Error: Not in dev environment. Cannot reset database."
		res.Header().Add("Content-Type", "text/plain; charset=utf-8")
		res.WriteHeader(http.StatusForbidden)
		res.Write([]byte(txt2))
	}

	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(txt))
}

func healthzHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("OK"))
}

func (cfg *apiConfig) createChirpHandler(res http.ResponseWriter, req *http.Request) {

	// need to decode the request body
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	// chirp validation logic
	if len(params.Body) > 140 {
		respondWithError(res, 400, "Chirp is too long")
		return
	}

	params.Body = replaceProfanity(params.Body)

	// need to encode the response body

	// validation complete, add chirp to chirps table in database
	chirp_params := database.CreateChirpParams{}
	chirp_params.Body = params.Body
	chirp_params.UserID = params.UserID

	// check that the user doesn't already exist
	chirp, err := cfg.db.CreateChirp(req.Context(), chirp_params)
	if err != nil {
		fmt.Printf("Error: couldn't create chirp '%s'.", chirp_params.Body)
		return
	}

	// need to encode the response body into a Chirp struct
	respBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(res, http.StatusCreated, respBody)

}

func (cfg *apiConfig) getChirpsHandler(res http.ResponseWriter, req *http.Request) {
	// returns all chirps in the database

	chirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		fmt.Print("Error: couldn't get chirps.")
		return
	}

	// need to encode the response body into a Chirps struct (need to change this so that it does the whole array, not a single chirp)

	respBody := []Chirp{}

	for _, chirp := range chirps {
		new_chirp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		respBody = append(respBody, new_chirp)
	}

	respondWithJSON(res, http.StatusOK, respBody)
}

func respondWithError(res http.ResponseWriter, code int, msg string) {
	type errorResp struct {
		Error string `json:"error"`
	}

	err := errorResp{Error: msg}

	respondWithJSON(res, code, err)
}

func respondWithJSON(res http.ResponseWriter, code int, payload any) {
	dat, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s", err)
		respondWithError(res, 500, err.Error())
		return
	}
	res.Header().Add("Content-Type", "application/json")
	res.WriteHeader(code)
	res.Write(dat)
}

func replaceProfanity(str string) string {
	//lower := strings.ToLower(str)
	words := strings.Split(str, " ")
	newWords := make([]string, 0) //len(words))
	for _, word := range words {
		if isProfane(word) {
			newWords = append(newWords, "****")
		} else {
			newWords = append(newWords, word)
		}
	}
	return strings.Join(newWords, " ")
}

func isProfane(str string) bool {
	lower := strings.ToLower(str)
	if lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
		return true
	}
	return false
}

func (cfg *apiConfig) createUserHandler(res http.ResponseWriter, req *http.Request) {
	// creates a user in the users database

	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	// check that the user doesn't already exist
	user, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		fmt.Printf("Error: couldn't create user %s. Possibly already exists.", params.Email)
		return
	}

	// need to encode the response body into a Users struct
	respBody := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(res, http.StatusCreated, respBody)

}

func main() {

	// get the info from the .env file
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)

	platform := os.Getenv("PLATFORM")

	// create the handler for the server
	cfg := apiConfig{}
	cfg.db = dbQueries
	cfg.platform = platform
	serve_mux := http.NewServeMux()
	serve_mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	serve_mux.HandleFunc("GET /api/healthz", healthzHandler)
	serve_mux.HandleFunc("GET /admin/metrics", cfg.printHitsHandler)
	serve_mux.HandleFunc("POST /admin/reset", cfg.resetHitsHandler)
	serve_mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	serve_mux.HandleFunc("GET /api/chirps", cfg.getChirpsHandler)
	serve_mux.HandleFunc("POST /api/users", cfg.createUserHandler)

	// create the server
	server_struct := http.Server{}
	server_struct.Addr = ":8080"
	server_struct.Handler = serve_mux
	err = server_struct.ListenAndServe()
	if err != nil {
		fmt.Print(err)
	}
}
