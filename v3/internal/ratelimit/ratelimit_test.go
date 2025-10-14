package ratelimit

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"

	"github.com/gobitfly/beaconchain-backend/internal/auth"
	"github.com/gobitfly/beaconchain-backend/internal/common"
	"github.com/gobitfly/beaconchain-backend/internal/domain"
	"github.com/gobitfly/beaconchain-backend/internal/limits"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// ------------------------------------------------
// Tests for isWithinRateLimit

func rateLimitTestSetup(t *testing.T) (context.Context, *miniredis.Miniredis, *redis.Client, *redis.Script) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return context.Background(), mr, client, redis.NewScript(scriptStr)
}

var limiter = limits.NewLimiter()

const testCallerA = "testCallerA"
const testCallerB = "testCallerB"

var testGlobalRateLimit, _ = limiter.GetRateLimit(context.Background(), domain.User{SubscriptionTier: domain.TierScale})
var testEndpointRateLimit, _ = limiter.GetRateLimit(context.Background(), domain.User{SubscriptionTier: domain.TierFree})

const testEndpointA = "/test/endpointA"
const testEndpointB = "/test/endpointB"

func TestIsWithinRateLimit_GlobalExceedBucketCapacity(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)

	now := time.Now()
	// Simulate requests within the burst limit
	for range testGlobalRateLimit.BucketCapacity {
		isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isWithinRateLimit, "Expected to be within rate limit")
	}
	// Simulate one more request which should exceed the rate limit
	isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
	assert.False(t, isWithinRateLimit, "Expected to be outside rate limit after 10 requests")
}

func TestIsWithinRateLimit_GlobalRefillSuccess(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)

	now := time.Now()
	// Simulate requests within the burst limit
	for range testGlobalRateLimit.BucketCapacity {
		isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isWithinRateLimit, "Expected to be within rate limit")
	}
	// Now wait for the refill interval to pass and make another request
	refillInterval := calcRefillInterval(testGlobalRateLimit.SteadyRate)
	now = now.Add(refillInterval)
	isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
	// After refill, we should be able to make another request
	assert.True(t, isWithinRateLimit, "Expected to be within rate limit after refill")
}
func TestIsWithinRateLimit_GlobalRefillSuccessAfterFixedInterval(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)

	now := time.Now()
	refillInterval := calcRefillInterval(testGlobalRateLimit.SteadyRate)
	fractionalRefillInterval := refillInterval / time.Duration(testGlobalRateLimit.BucketCapacity)
	// Simulate requests within the burst limit spread evenly over one refill interval
	for range testGlobalRateLimit.BucketCapacity - 1 {
		isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isWithinRateLimit, "Expected to be within rate limit")
		now = now.Add(fractionalRefillInterval) // increment time by fractional refill interval
	}
	now = now.Add(fractionalRefillInterval - time.Millisecond) // ensure we are just before the refill
	isInRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
	assert.True(t, isInRateLimit, "Expected to be within rate limit after refill")

	// after waiting for ~n+1 fractional refill intervals, exactly one full refill interval should have passed, so we should be able to make another request
	now = now.Add(fractionalRefillInterval)
	isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
	assert.True(t, isInRateLimit, "Expected to be outside rate limit after exceeding bucket capacity")

	// after waiting for another fractional refill interval, we should be outside the rate limit
	now = now.Add(fractionalRefillInterval)
	isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
	assert.False(t, isInRateLimit, "Expected to be outside rate limit after exceeding bucket capacity")
}

func TestIsWithinRateLimit_GlobalRefillBurst(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)

	now := time.Now()
	// Simulate 10 requests within the rate limit
	for range testGlobalRateLimit.BucketCapacity {
		isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isWithinRateLimit, "Expected to be within rate limit")
	}
	// Now wait for enough time to refill the bucket and burst again
	refillInterval := calcRefillInterval(testGlobalRateLimit.SteadyRate)
	now = now.Add(refillInterval * time.Duration(testGlobalRateLimit.BucketCapacity))
	for range testGlobalRateLimit.BucketCapacity {
		isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isWithinRateLimit, "Expected to be within rate limit")
	}
	// Simulate one more request which should exceed the rate limit
	isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
	assert.False(t, isWithinRateLimit, "Expected to be outside rate limit after 10 requests")
}

func TestIsWithinRateLimit_GlobalRefillAtSteadyRate(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)

	now := time.Now()
	refillInterval := calcRefillInterval(testGlobalRateLimit.SteadyRate)
	// Simulate requests within the steady rate limit that would otherwise exceed the burst limit
	for range 10 * testGlobalRateLimit.BucketCapacity {
		isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isWithinRateLimit, "Expected to be within rate limit")
		now = now.Add(refillInterval)
	}
}

func TestIsWithinRateLimit_EndpointExceedsBeforeGlobal(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)

	globalRateLimit := testGlobalRateLimit
	endpointRateLimit := testEndpointRateLimit
	assert.Greater(t, globalRateLimit.BucketCapacity, endpointRateLimit.BucketCapacity, "Endpoint rate limit should have a larger bucket capacity than global rate limit")
	assert.Greater(t, globalRateLimit.SteadyRate, endpointRateLimit.SteadyRate, "Endpoint rate limit should have a larger steady rate than global rate limit")
	now := time.Now()
	// Simulate requests within the endpoint burst limit
	for range endpointRateLimit.BucketCapacity {
		isWithinRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, globalRateLimit, endpointRateLimit)
		assert.True(t, isWithinRateLimit, "Expected to be within rate limit")
	}
	// Now make one more request which should exceed the endpoint rate limit
	isInRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, globalRateLimit, endpointRateLimit)
	assert.False(t, isInRateLimit, "Expected to be outside rate limit after exceeding endpoint burst limit")

	// Now make a request to endpoint B which should still be within the global rate limit
	isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointB, globalRateLimit, endpointRateLimit)
	assert.True(t, isInRateLimit, "Expected to be within rate limit for endpoint B after exceeding endpoint A burst limit")
}

func TestIsWithinRateLimit_DifferentCallersDoNotShareBuckets(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)
	now := time.Now()
	for range testGlobalRateLimit.BucketCapacity {
		isInRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isInRateLimit, "Expected caller A to be within rate limit")
		isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerB, testEndpointA, testGlobalRateLimit, testGlobalRateLimit)
		assert.True(t, isInRateLimit, "Expected caller B to be within rate limit")
	}
}

func TestIsWithinRateLimit_FailsOpen(t *testing.T) {
	ctx, _, client, script := rateLimitTestSetup(t)
	now := time.Now()
	testRateLimit := &limits.RateLimitSettings{
		BucketCapacity: 1,
		SteadyRate:     1,
	}
	// exhaust the bucket
	isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testRateLimit, testRateLimit)
	isInRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testRateLimit, testRateLimit)
	assert.False(t, isInRateLimit, "Expected caller A to be outside rate limit due to zero bucket capacity and steady rate")

	// Simulate a Redis outage by making closing the client
	err := client.Close()
	assert.NoError(t, err, "Expected to close Redis client without error")
	// Now the caller should still be able to make requests, as we fail open
	isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testRateLimit, testRateLimit)
	assert.True(t, isInRateLimit, "Expected caller A to be within rate limit due to Redis outage (fail open)")
}

func TestIsWithinRateLimit_SetsExpirationCorrectly(t *testing.T) {
	ctx, mr, client, script := rateLimitTestSetup(t)

	now := time.Now()
	testRateLimit := &limits.RateLimitSettings{
		BucketCapacity: 10,
		SteadyRate:     0.1,
	}
	isInRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testRateLimit, testRateLimit)
	assert.True(t, isInRateLimit, "Expected caller A to be within rate limit")
	// waiting for the refill interval to pass inside redis should expire the key
	refillInterval := calcRefillInterval(testRateLimit.SteadyRate)
	mr.FastForward(refillInterval)
	key := getCallerKey(testCallerA)
	_, err := mr.Get(key)
	assert.ErrorIs(t, err, miniredis.ErrKeyNotFound, "Expected key to be expired after refill interval")

	// expiration should be set to refillInterval*requestCount (as long as no refill happened)
	reqCount := testRateLimit.BucketCapacity / 2
	for range reqCount {
		isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testRateLimit, testRateLimit)
		assert.True(t, isInRateLimit, "Expected caller A to be within rate limit")
	}
	expectedTTL := refillInterval * time.Duration(reqCount)
	ttl := mr.TTL(key)
	assert.Equal(t, expectedTTL, ttl, "Expected TTL to be set correctly after burst requests")
}

func TestIsWithinRateLimit_SetsExpirationCorrectlyForMultipleEndpoints(t *testing.T) {
	ctx, mr, client, script := rateLimitTestSetup(t)

	now := time.Now()
	testRateLimitA := &limits.RateLimitSettings{
		BucketCapacity: 10,
		SteadyRate:     0.1,
	}
	testRateLimitB := &limits.RateLimitSettings{
		BucketCapacity: testRateLimitA.BucketCapacity,
		SteadyRate:     testRateLimitA.SteadyRate / 10, // lower steady rate should refill slower -> higher TTL
	}
	isInRateLimit := isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testRateLimitA, testRateLimitA) // simulate 1 request with higher steady rate
	assert.True(t, isInRateLimit, "Expected caller A to be within rate limit")
	expectedTTLA := calcRefillInterval(testRateLimitA.SteadyRate)
	key := getCallerKey(testCallerA)
	ttl := mr.TTL(key)
	assert.Equal(t, expectedTTLA, ttl, "Expected TTL to be set correctly")

	isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointB, testRateLimitA, testRateLimitB) // simulate 1 request with lower steady rate
	assert.True(t, isInRateLimit, "Expected caller A to be within rate limit for endpoint B")
	// TTL for endpoint B should be greater than for endpoint A due to lower steady rate
	expectedTTLB := calcRefillInterval(testRateLimitB.SteadyRate)
	assert.Greater(t, expectedTTLB, expectedTTLA, "Expected TTL for endpoint B to be greater than endpoint A due to lower steady rate")
	ttl = mr.TTL(key)
	assert.Equal(t, expectedTTLB, ttl, "Expected TTL to be set correctly for endpoint B")

	// simulating another request to endpoint A should not change the TTL, as it has a higher steady rate
	isInRateLimit = isWithinRateLimit(ctx, client, script, now, testCallerA, testEndpointA, testRateLimitA, testRateLimitA)
	assert.True(t, isInRateLimit, "Expected caller A to be within rate limit for endpoint A")
	ttl = mr.TTL(key)
	assert.Equal(t, expectedTTLB, ttl, "Expected TTL to remain the same after another request to endpoint A")
}

func TestCalcRefillInterval(t *testing.T) {
	tests := []struct {
		name       string
		steadyRate float32
		expected   time.Duration
	}{
		{"Zero Steady Rate", 0, 0},
		{"One Steady Rate", 1, time.Second},
		{"Two Steady Rate", 2, time.Millisecond * 500},
		{"Ten Steady Rate", 10, time.Millisecond * 100},
		{"Fractional Steady Rate", 0.5, time.Second * 2},
		{"Fractional Steady Rate #2", 2.5, time.Millisecond * 400},
		{"Large Steady Rate", 100, time.Millisecond * 10},
		{"Very Small Steady Rate", 0.01, time.Second * 100},
		{"Negative Steady Rate", -1, 0}, // Negative rates should return 0 duration
		{"Rate Approaches Zero (Small Positive)", 0.0000001, time.Duration(1.0 / 0.0000001 * float64(time.Second))},
		{"Rate Is Just Above Zero", 0.000001, time.Second * 1000000},
		{"Rate Is Very High", 1_000_000, time.Microsecond},
		{"Rate Causing Truncation to Zero (Very High)", 1_000_000_000_000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calcRefillInterval(tt.steadyRate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ------------------------------------------------
// Tests for GetRateLimitMiddleware

var testHandler = func(_ context.Context, _ any) (any, error) {
	return "ok", nil
}

type testScripterStub struct {
	isRequestAllowed bool
}

func newTestScripterStub(isRequestAllowed bool) *testScripterStub {
	return &testScripterStub{isRequestAllowed: isRequestAllowed}
}
func (s *testScripterStub) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	var val int64 = 0
	if s.isRequestAllowed {
		val = 1
	}
	cmd := redis.NewCmd(ctx)
	cmd.SetVal(val)
	return cmd
}
func (s *testScripterStub) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return s.Eval(ctx, "", keys, args...)
}
func (s *testScripterStub) ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd {
	return redis.NewBoolSliceCmd(ctx)
}
func (s *testScripterStub) ScriptLoad(ctx context.Context, script string) *redis.StringCmd {
	return redis.NewStringCmd(ctx)
}

var _ redis.Scripter = (*testScripterStub)(nil) // Ensure testScripterStub implements redis.Scripter

func testGetEndpointRatelinit(fullMethod string, tier domain.Tier) (*limits.RateLimitSettings, error) {
	return &limits.RateLimitSettings{
		BucketCapacity: 5,
		SteadyRate:     1,
	}, nil
}

var testServerInfo = &grpc.UnaryServerInfo{}
var testUser = domain.User{
	ID:               123,
	SubscriptionTier: domain.TierFree,
}

func TestGetRateLimitMiddleware_SuccessWithUserInContext(t *testing.T) {
	scripter := newTestScripterStub(true /* isRequestAllowed */)
	middleware := GetRateLimitMiddleware(scripter, testGetEndpointRatelinit)

	ctx := auth.SetUserInContext(context.Background(), testUser)

	response, err := middleware(ctx, nil, testServerInfo, testHandler)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response)
}

func TestGetRateLimitMiddleware_SuccessfulRatelimit(t *testing.T) {
	scripter := newTestScripterStub(false /* isRequestAllowed */)
	middleware := GetRateLimitMiddleware(scripter, testGetEndpointRatelinit)

	ctx := auth.SetUserInContext(context.Background(), testUser)

	_, err := middleware(ctx, nil, testServerInfo, testHandler)
	assert.Error(t, err)
	assert.Equal(t, common.Code(err), http.StatusTooManyRequests)
}

func TestGetRateLimitMiddleware_ErrorWithoutUserInContext(t *testing.T) {
	scripter := newTestScripterStub(false /* isRequestAllowed */)
	middleware := GetRateLimitMiddleware(scripter, testGetEndpointRatelinit)

	ctx := context.Background()

	_, err := middleware(ctx, nil, testServerInfo, testHandler)
	assert.Error(t, err)
	assert.Equal(t, common.Code(err), http.StatusInternalServerError)
}

func TestGetRateLimitMiddleware_SuccessWithNilRatelimitOpts(t *testing.T) {
	scripter := newTestScripterStub(false /* isRequestAllowed */)
	getRatelimitOpts := func(fullMethod string, tier domain.Tier) (*limits.RateLimitSettings, error) {
		return nil, nil // should cause fallback to global ratelimit
	}
	middleware := GetRateLimitMiddleware(scripter, getRatelimitOpts)

	ctx := auth.SetUserInContext(context.Background(), testUser)

	_, err := middleware(ctx, nil, testServerInfo, testHandler)
	assert.Error(t, err)
	assert.Equal(t, common.Code(err), http.StatusTooManyRequests)
}

func TestGetRateLimitMiddleware_ErrorRateLimitOpts(t *testing.T) {
	scripter := newTestScripterStub(false /* isRequestAllowed */)
	getRatelimitOpts := func(fullMethod string, tier domain.Tier) (*limits.RateLimitSettings, error) {
		return nil, common.NewAPIInternalError(http.StatusInternalServerError, "test error getting ratelimit opts")
	}
	middleware := GetRateLimitMiddleware(scripter, getRatelimitOpts)

	ctx := auth.SetUserInContext(context.Background(), testUser)

	_, err := middleware(ctx, nil, testServerInfo, testHandler)
	assert.Equal(t, common.Code(err), http.StatusInternalServerError)
}
