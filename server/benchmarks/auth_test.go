package benchmarks

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func BenchmarkBcryptHash(b *testing.B) {
	password := []byte("benchmark-password-123!")
	for _, cost := range []int{10, 12, 14} {
		b.Run(fmt.Sprintf("cost%d", cost), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bcrypt.GenerateFromPassword(password, cost)
			}
		})
	}
}

func BenchmarkBcryptCompare(b *testing.B) {
	password := []byte("benchmark-password-123!")
	hash, _ := bcrypt.GenerateFromPassword(password, 12)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bcrypt.CompareHashAndPassword(hash, password)
	}
}

func BenchmarkBcryptCompare_Parallel(b *testing.B) {
	password := []byte("benchmark-password-123!")
	hash, _ := bcrypt.GenerateFromPassword(password, 12)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bcrypt.CompareHashAndPassword(hash, password)
		}
	})
}

func BenchmarkJWTSign_HS512(b *testing.B) {
	secret := []byte("benchmark-test-secret-256-bits-long-enough")
	userID := uuid.New()
	sessionID := uuid.New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		claims := jwt.MapClaims{
			"uid": userID.String(),
			"sid": sessionID.String(),
			"exp": time.Now().Add(15 * time.Minute).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		token.SignedString(secret)
	}
}

func BenchmarkJWTValidate_HS512(b *testing.B) {
	secret := []byte("benchmark-test-secret-256-bits-long-enough")
	claims := jwt.MapClaims{
		"uid": uuid.New().String(),
		"sid": uuid.New().String(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, _ := token.SignedString(secret)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})
	}
}

func BenchmarkJWTSignAndValidate(b *testing.B) {
	secret := []byte("benchmark-test-secret-256-bits-long-enough")
	userID := uuid.New()
	sessionID := uuid.New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		claims := jwt.MapClaims{
			"uid": userID.String(),
			"sid": sessionID.String(),
			"exp": time.Now().Add(15 * time.Minute).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenStr, _ := token.SignedString(secret)
		jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})
	}
}

func BenchmarkRefreshTokenGen(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 48)
		rand.Read(buf)
		base64.URLEncoding.EncodeToString(buf)
	}
}

func BenchmarkSHA256(b *testing.B) {
	data := []byte("benchmark-token-data-for-node-auth-hashing")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sha256.Sum256(data)
	}
}

func BenchmarkUUIDGeneration(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uuid.New()
	}
}
