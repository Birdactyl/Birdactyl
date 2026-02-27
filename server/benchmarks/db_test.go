package benchmarks

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"birdactyl-panel-backend/internal/database"
	"birdactyl-panel-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func BenchmarkDBUserCreate(b *testing.B) {
	requireDB(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := uuid.New().String()[:8]
		user := &models.User{
			Email:        fmt.Sprintf("bench_create_%s@test.com", uid),
			Username:     fmt.Sprintf("bench_create_%s", uid),
			PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
			RegisterIP:   "127.0.0.1",
		}
		database.DB.Create(user)
	}
}

func BenchmarkDBUserLookupByEmail(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_email_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_email_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.User
		database.DB.Where("email = ?", user.Email).First(&found)
	}
}

func BenchmarkDBUserLookupByID(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_id_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_id_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.User
		database.DB.Where("id = ?", user.ID).First(&found)
	}
}

func BenchmarkDBSessionCreate(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_sess_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_sess_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := &models.Session{
			UserID:       user.ID,
			RefreshToken: fmt.Sprintf("rt_%d_%s", i, uuid.New().String()),
			UserAgent:    "BenchmarkAgent/1.0",
			IP:           "127.0.0.1",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		database.DB.Create(session)
	}
}

func BenchmarkDBSessionLookup(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_slook_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_slook_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	session := &models.Session{
		UserID:       user.ID,
		RefreshToken: fmt.Sprintf("rt_lookup_%s", uuid.New().String()),
		UserAgent:    "BenchmarkAgent/1.0",
		IP:           "127.0.0.1",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	database.DB.Create(session)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.Session
		database.DB.Where("refresh_token = ?", session.RefreshToken).First(&found)
	}
}

func BenchmarkDBSessionCount(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_scount_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_scount_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)
	for j := 0; j < 25; j++ {
		database.DB.Create(&models.Session{
			UserID:       user.ID,
			RefreshToken: fmt.Sprintf("rt_count_%d_%s", j, uuid.New().String()),
			UserAgent:    "BenchmarkAgent/1.0",
			IP:           "127.0.0.1",
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var count int64
		database.DB.Model(&models.Session{}).Where("user_id = ?", user.ID).Count(&count)
	}
}

func BenchmarkDBCleanExpiredSessions(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_clean_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_clean_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for j := 0; j < 10; j++ {
			database.DB.Create(&models.Session{
				UserID:       user.ID,
				RefreshToken: fmt.Sprintf("rt_exp_%d_%d_%s", i, j, uuid.New().String()),
				UserAgent:    "BenchmarkAgent/1.0",
				IP:           "127.0.0.1",
				ExpiresAt:    time.Now().Add(-1 * time.Hour),
			})
		}
		b.StartTimer()
		database.DB.Where("expires_at < ?", time.Now()).Delete(&models.Session{})
	}
}

func BenchmarkDBConcurrentReads(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_cread_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_cread_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var found models.User
			database.DB.Where("id = ?", user.ID).First(&found)
		}
	})
}

func BenchmarkDBConcurrentWrites(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_cwrite_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_cwrite_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	var counter int64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddInt64(&counter, 1)
			session := &models.Session{
				UserID:       user.ID,
				RefreshToken: fmt.Sprintf("rt_cw_%d_%s", n, uuid.New().String()),
				UserAgent:    "BenchmarkAgent/1.0",
				IP:           "127.0.0.1",
				ExpiresAt:    time.Now().Add(24 * time.Hour),
			}
			database.DB.Create(session)
		}
	})
}

func BenchmarkDBConcurrentMixed(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_cmix_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_cmix_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	var counter int64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n := atomic.AddInt64(&counter, 1)
			if n%3 == 0 {
				session := &models.Session{
					UserID:       user.ID,
					RefreshToken: fmt.Sprintf("rt_cm_%d_%s", n, uuid.New().String()),
					UserAgent:    "BenchmarkAgent/1.0",
					IP:           "127.0.0.1",
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				}
				database.DB.Create(session)
			} else {
				var found models.User
				database.DB.Where("id = ?", user.ID).First(&found)
			}
		}
	})
}

func BenchmarkDBTransaction(b *testing.B) {
	requireDB(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := uuid.New().String()[:8]
		database.DB.Transaction(func(tx *gorm.DB) error {
			user := &models.User{
				Email:        fmt.Sprintf("bench_tx_%s@test.com", uid),
				Username:     fmt.Sprintf("bench_tx_%s", uid),
				PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
				RegisterIP:   "127.0.0.1",
			}
			if err := tx.Create(user).Error; err != nil {
				return err
			}
			return tx.Create(&models.IPRegistration{IP: "127.0.0.1", UserID: user.ID}).Error
		})
	}
}
