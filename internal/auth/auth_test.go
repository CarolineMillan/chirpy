package auth

import (
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	// 1. pick a plaintext password string, e.g. "correctPassword123!"
	// 2. call HashPassword with it
	// 3. check err is nil
	// 4. check the returned hash is NOT equal to the original password
	//    (a hash should never look like the plaintext)
	// 5. optional: check the hash isn't an empty string
}

func TestCheckPasswordHash(t *testing.T) {
	// 1. pick a plaintext password, hash it once using HashPassword to get a valid hash
	// 2. define test cases such as:
	//    - correct password + its real hash -> expect match=true, err=nil
	//    - wrong password + that hash       -> expect match=false, err=nil
	//    - correct password + garbage string as "hash" -> expect an error
	// 3. loop through cases, calling CheckPasswordHash(password, hash) for each
	// 4. compare the actual (match, err) against what you expected for that case
}

func TestMakeAndValidateJWT(t *testing.T) {
	// 1. create a uuid to act as your test user's ID
	test_uuid, err := uuid.NewUUID()

	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	// 2. call MakeJWT with that ID, a test secret, and some expiration duration
	token, err := MakeJWT(test_uuid, "my-secret-key", time.Minute)

	// 3. check err is nil
	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	// 4. call ValidateJWT with the token and the *same* secret
	new_uuid, err := ValidateJWT(token, "my-secret-key")

	// 5. check err is nil
	if err != nil {
		t.Fatalf("ValidateJWT() returned an error: %v", err)
	}

	// 6. check the returned uuid matches the one you started with
	if test_uuid != new_uuid {
		t.Fatalf("expected %v, got %v", test_uuid, new_uuid)
	}

	// pass
}

func TestValidateJWTExpired(t *testing.T) {
	// 1. create a test uuid
	test_uuid, err := uuid.NewUUID()

	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	// 2. call MakeJWT, but pass a NEGATIVE duration (e.g. -time.Hour)
	//    this makes the token's expiry time already be in the past
	token, err := MakeJWT(test_uuid, "my-secret-key", -time.Minute)

	// 3. check err is nil (MakeJWT itself should still succeed)
	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	// 4. call ValidateJWT with that token and the same secret
	_, err = ValidateJWT(token, "my-secret-key")

	// 5. this time, you EXPECT an error - check that err is NOT nil
	//    (if err IS nil, that's the failure case - use t.Errorf)
	if err == nil {
		t.Fatalf("ValidateJWT() returned an error: %v", err)
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	// 1. create a test uuid
	test_uuid, err := uuid.NewUUID()

	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	// 2. call MakeJWT with a valid duration and secret "correct-secret"
	token, err := MakeJWT(test_uuid, "correct-secret", time.Minute)

	// 3. check err is nil
	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}

	// 4. call ValidateJWT with the token, but pass "wrong-secret" instead
	_, err = ValidateJWT(token, "wrong-secret")

	// 5. you EXPECT an error here too - check err is NOT nil
	if err == nil {
		t.Fatalf("ValidateJWT() returned an error: %v", err)
	}
}
