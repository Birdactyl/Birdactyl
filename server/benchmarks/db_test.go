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

func benchSetupNodeAndPackage(b *testing.B) (uuid.UUID, uuid.UUID) {
	b.Helper()
	uid := uuid.New().String()[:8]
	node := &models.Node{
		Name:      fmt.Sprintf("bench_node_%s", uid),
		FQDN:      fmt.Sprintf("node-%s.bench.local", uid),
		Port:      8443,
		TokenID:   fmt.Sprintf("tid_%s", uid),
		TokenHash: "bench_hash",
		DaemonToken: "bench_daemon",
		IsOnline:  true,
	}
	database.DB.Create(node)

	pkg := &models.Package{
		Name:        fmt.Sprintf("bench_pkg_%s", uid),
		DockerImage: "ghcr.io/test/bench:latest",
		Startup:     "java -jar server.jar",
	}
	database.DB.Create(pkg)
	return node.ID, pkg.ID
}

func BenchmarkDBServerCreate(b *testing.B) {
	requireDB(b)
	nodeID, pkgID := benchSetupNodeAndPackage(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_srv_create_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_srv_create_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server := &models.Server{
			Name:      fmt.Sprintf("bench_srv_%d", i),
			UserID:    user.ID,
			NodeID:    nodeID,
			PackageID: pkgID,
			Status:    models.ServerStatusInstalling,
			Memory:    1024,
			CPU:       100,
			Disk:      5120,
			Ports:     []byte(`[{"port":25565,"primary":true}]`),
			Variables: []byte(`{}`),
		}
		database.DB.Create(server)
	}
}

func BenchmarkDBServerLookupByID(b *testing.B) {
	requireDB(b)
	nodeID, pkgID := benchSetupNodeAndPackage(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_srv_id_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_srv_id_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)
	server := &models.Server{
		Name: "bench_lookup_srv", UserID: user.ID, NodeID: nodeID, PackageID: pkgID,
		Status: models.ServerStatusRunning, Memory: 1024, CPU: 100, Disk: 5120,
		Ports: []byte(`[{"port":25565,"primary":true}]`), Variables: []byte(`{}`),
	}
	database.DB.Create(server)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.Server
		database.DB.Where("id = ?", server.ID).First(&found)
	}
}

func BenchmarkDBServerLookupByUser(b *testing.B) {
	requireDB(b)
	nodeID, pkgID := benchSetupNodeAndPackage(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_srv_user_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_srv_user_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)
	for j := 0; j < 10; j++ {
		database.DB.Create(&models.Server{
			Name: fmt.Sprintf("bench_srv_%d", j), UserID: user.ID, NodeID: nodeID, PackageID: pkgID,
			Status: models.ServerStatusRunning, Memory: 1024, CPU: 100, Disk: 5120,
			Ports: []byte(`[{"port":25565,"primary":true}]`), Variables: []byte(`{}`),
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var servers []models.Server
		database.DB.Where("user_id = ?", user.ID).Find(&servers)
	}
}

func BenchmarkDBServerPreload(b *testing.B) {
	requireDB(b)
	nodeID, pkgID := benchSetupNodeAndPackage(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_srv_preload_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_srv_preload_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)
	server := &models.Server{
		Name: "bench_preload_srv", UserID: user.ID, NodeID: nodeID, PackageID: pkgID,
		Status: models.ServerStatusRunning, Memory: 1024, CPU: 100, Disk: 5120,
		Ports: []byte(`[{"port":25565,"primary":true}]`), Variables: []byte(`{}`),
	}
	database.DB.Create(server)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.Server
		database.DB.Preload("User").Preload("Node").Preload("Package").Where("id = ?", server.ID).First(&found)
	}
}

func BenchmarkDBNodeCreate(b *testing.B) {
	requireDB(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uid := uuid.New().String()[:8]
		node := &models.Node{
			Name:        fmt.Sprintf("bench_node_%s", uid),
			FQDN:        fmt.Sprintf("node-%s.bench.local", uid),
			Port:        8443,
			TokenID:     fmt.Sprintf("tid_%s_%d", uid, i),
			TokenHash:   "bench_hash",
			DaemonToken: "bench_daemon",
			IsOnline:    true,
		}
		database.DB.Create(node)
	}
}

func BenchmarkDBNodeLookup(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	node := &models.Node{
		Name:        fmt.Sprintf("bench_nlook_%s", uid),
		FQDN:        fmt.Sprintf("nlook-%s.bench.local", uid),
		Port:        8443,
		TokenID:     fmt.Sprintf("tid_nlook_%s", uid),
		TokenHash:   "bench_hash",
		DaemonToken: "bench_daemon",
		IsOnline:    true,
	}
	database.DB.Create(node)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.Node
		database.DB.Where("id = ?", node.ID).First(&found)
	}
}

func BenchmarkDBActivityLogInsert(b *testing.B) {
	requireDB(b)
	userID := uuid.New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log := &models.ActivityLog{
			UserID:      userID,
			Username:    "benchuser",
			Action:      "auth.login",
			Description: "User logged in",
			IP:          "127.0.0.1",
			UserAgent:   "BenchmarkAgent/1.0",
			IsAdmin:     false,
		}
		database.DB.Create(log)
	}
}

func BenchmarkDBActivityLogQuery(b *testing.B) {
	requireDB(b)
	userID := uuid.New()
	for j := 0; j < 50; j++ {
		database.DB.Create(&models.ActivityLog{
			UserID:      userID,
			Username:    "benchuser",
			Action:      "auth.login",
			Description: "User logged in",
			IP:          "127.0.0.1",
			UserAgent:   "BenchmarkAgent/1.0",
			IsAdmin:     false,
		})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var logs []models.ActivityLog
		database.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(20).Find(&logs)
	}
}

func BenchmarkDBAPIKeyCreate(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_apikey_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_apikey_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := &models.APIKey{
			UserID:    user.ID,
			Name:      fmt.Sprintf("bench_key_%d", i),
			KeyHash:   fmt.Sprintf("hash_%d_%s", i, uuid.New().String()),
			KeyPrefix: "birdactyl_",
		}
		database.DB.Create(key)
	}
}

func BenchmarkDBAPIKeyLookup(b *testing.B) {
	requireDB(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_aklook_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_aklook_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)
	key := &models.APIKey{
		UserID:    user.ID,
		Name:      "bench_lookup_key",
		KeyHash:   fmt.Sprintf("hash_lookup_%s", uuid.New().String()),
		KeyPrefix: "birdactyl_",
	}
	database.DB.Create(key)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.APIKey
		database.DB.Where("id = ?", key.ID).First(&found)
	}
}

func BenchmarkDBSubuserCreate(b *testing.B) {
	requireDB(b)
	nodeID, pkgID := benchSetupNodeAndPackage(b)
	uid := uuid.New().String()[:8]
	owner := &models.User{
		Email:        fmt.Sprintf("bench_subowner_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_subowner_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(owner)
	server := &models.Server{
		Name: "bench_sub_srv", UserID: owner.ID, NodeID: nodeID, PackageID: pkgID,
		Status: models.ServerStatusRunning, Memory: 1024, CPU: 100, Disk: 5120,
		Ports: []byte(`[{"port":25565,"primary":true}]`), Variables: []byte(`{}`),
	}
	database.DB.Create(server)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subUID := uuid.New().String()[:8]
		subuser := &models.User{
			Email:        fmt.Sprintf("bench_sub_%s_%d@test.com", subUID, i),
			Username:     fmt.Sprintf("bench_sub_%s_%d", subUID, i),
			PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
			RegisterIP:   "127.0.0.1",
		}
		database.DB.Create(subuser)

		sub := &models.Subuser{
			ServerID:    server.ID,
			UserID:      subuser.ID,
			Permissions: []byte(`["power.start","console.read","file.list"]`),
		}
		database.DB.Create(sub)
	}
}

func BenchmarkDBSubuserLookup(b *testing.B) {
	requireDB(b)
	nodeID, pkgID := benchSetupNodeAndPackage(b)
	uid := uuid.New().String()[:8]
	owner := &models.User{
		Email:        fmt.Sprintf("bench_sublook_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_sublook_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(owner)
	server := &models.Server{
		Name: "bench_sublook_srv", UserID: owner.ID, NodeID: nodeID, PackageID: pkgID,
		Status: models.ServerStatusRunning, Memory: 1024, CPU: 100, Disk: 5120,
		Ports: []byte(`[{"port":25565,"primary":true}]`), Variables: []byte(`{}`),
	}
	database.DB.Create(server)

	subuser := &models.User{
		Email:        fmt.Sprintf("bench_sublook_u_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_sublook_u_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(subuser)
	sub := &models.Subuser{
		ServerID:    server.ID,
		UserID:      subuser.ID,
		Permissions: []byte(`["power.start","console.read","file.list"]`),
	}
	database.DB.Create(sub)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var found models.Subuser
		database.DB.Where("server_id = ? AND user_id = ?", server.ID, subuser.ID).First(&found)
	}
}

func BenchmarkDBServerUpdate(b *testing.B) {
	requireDB(b)
	nodeID, pkgID := benchSetupNodeAndPackage(b)
	uid := uuid.New().String()[:8]
	user := &models.User{
		Email:        fmt.Sprintf("bench_srv_upd_%s@test.com", uid),
		Username:     fmt.Sprintf("bench_srv_upd_%s", uid),
		PasswordHash: "$2a$12$000000000000000000000.0000000000000000000000000000000",
		RegisterIP:   "127.0.0.1",
	}
	database.DB.Create(user)
	server := &models.Server{
		Name: "bench_update_srv", UserID: user.ID, NodeID: nodeID, PackageID: pkgID,
		Status: models.ServerStatusRunning, Memory: 1024, CPU: 100, Disk: 5120,
		Ports: []byte(`[{"port":25565,"primary":true}]`), Variables: []byte(`{}`),
	}
	database.DB.Create(server)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		database.DB.Model(&models.Server{}).Where("id = ?", server.ID).Updates(map[string]interface{}{
			"status": models.ServerStatusRunning,
			"memory": 2048 + (i % 4096),
		})
	}
}

