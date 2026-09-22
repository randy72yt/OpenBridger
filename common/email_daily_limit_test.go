package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// The daily cap exists to protect the upstream provider quota (Resend free tier
// is 100/day). These tests pin the parts that do not need a live Redis.

func TestSecondsUntilTomorrowStaysWithinOneDay(t *testing.T) {
	v := secondsUntilTomorrow()
	require.Greater(t, v, int64(0))
	require.LessOrEqual(t, v, int64(24*60*60+60))
}

func TestEmailDailyQuotaKeyUsesCurrentDate(t *testing.T) {
	require.Regexp(t, `^email:daily:\d{4}-\d{2}-\d{2}$`, emailDailyQuotaKey())
}

func TestTakeEmailDailyQuotaDisabledByConfig(t *testing.T) {
	origEnable, origNum := EmailDailyLimitEnable, EmailDailyLimitNum
	t.Cleanup(func() { EmailDailyLimitEnable, EmailDailyLimitNum = origEnable, origNum })

	EmailDailyLimitEnable = false
	EmailDailyLimitNum = 0
	require.NoError(t, takeEmailDailyQuota())
}

// Without Redis there is no counter to consult, so sending must not be blocked.
// Losing verification mail entirely would be worse than losing the accounting.
func TestTakeEmailDailyQuotaFailsOpenWithoutRedis(t *testing.T) {
	origEnable, origNum, origRedis := EmailDailyLimitEnable, EmailDailyLimitNum, RedisEnabled
	t.Cleanup(func() {
		EmailDailyLimitEnable, EmailDailyLimitNum, RedisEnabled = origEnable, origNum, origRedis
	})

	EmailDailyLimitEnable = true
	EmailDailyLimitNum = 1
	RedisEnabled = false
	require.NoError(t, takeEmailDailyQuota())
}

func TestTakeEmailDailyQuotaFailsOpenWithNilClient(t *testing.T) {
	origEnable, origNum, origRedis := EmailDailyLimitEnable, EmailDailyLimitNum, RedisEnabled
	t.Cleanup(func() {
		EmailDailyLimitEnable, EmailDailyLimitNum, RedisEnabled = origEnable, origNum, origRedis
	})

	EmailDailyLimitEnable = true
	EmailDailyLimitNum = 1
	RedisEnabled = true
	RDB = nil
	require.NoError(t, takeEmailDailyQuota())
}
