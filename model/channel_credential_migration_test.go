package model

import (
	"fmt"
	"os"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// legacyChannel is the channel row shape from a binary that does not know the
// upstream access-token columns. Inserts from that binary must still succeed
// after AutoMigrate adds the columns.
type legacyChannel struct {
	ID     int    `gorm:"primaryKey"`
	Key    string `gorm:"column:key;not null"`
	Name   string
	Models string
}

func (legacyChannel) TableName() string { return "channels" }

func TestChannelCredentialColumnsMigrate(t *testing.T) {
	cases := []struct {
		name string
		open func() gorm.Dialector
	}{
		{
			name: "sqlite",
			open: func() gorm.Dialector {
				return sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()))
			},
		},
	}
	if dsn := os.Getenv("TEST_MYSQL_DSN"); dsn != "" {
		cases = append(cases, struct {
			name string
			open func() gorm.Dialector
		}{name: "mysql", open: func() gorm.Dialector { return mysql.Open(dsn) }})
	}
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		cases = append(cases, struct {
			name string
			open func() gorm.Dialector
		}{name: "postgres", open: func() gorm.Dialector {
			return postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
		}})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(tc.open(), &gorm.Config{})
			require.NoError(t, err)
			sqlDB, err := db.DB()
			require.NoError(t, err)
			t.Cleanup(func() { _ = sqlDB.Close() })

			if tc.name != "sqlite" {
				require.NoError(t, db.Migrator().DropTable(&Channel{}))
			}
			require.NoError(t, db.AutoMigrate(&Channel{}))
			require.NoError(t, db.AutoMigrate(&Channel{}))
			require.True(t, db.Migrator().HasColumn(&Channel{}, "upstream_access_token"))
			require.True(t, db.Migrator().HasColumn(&Channel{}, "upstream_user_id"))

			legacy := legacyChannel{Key: "sk-old", Name: "old-binary-" + tc.name, Models: "gpt-5,gpt-5"}
			require.NoError(t, db.Create(&legacy).Error)

			var loaded Channel
			require.NoError(t, db.First(&loaded, legacy.ID).Error)
			require.Empty(t, loaded.UpstreamAccessToken)
			require.Equal(t, 0, loaded.UpstreamUserID)
			require.False(t, loaded.UpstreamAccessTokenConfigured)

			require.NoError(t, db.Model(&loaded).Updates(map[string]any{
				"upstream_access_token": "stored-token",
				"upstream_user_id":      9,
			}).Error)
			var roundTrip Channel
			require.NoError(t, db.First(&roundTrip, legacy.ID).Error)
			require.Equal(t, "stored-token", roundTrip.UpstreamAccessToken)
			require.Equal(t, 9, roundTrip.UpstreamUserID)
			encoded, err := common.Marshal(roundTrip)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "stored-token")
		})
	}
}
