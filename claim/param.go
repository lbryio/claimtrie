package claim

// Param ...
type Param func()

// SetParams ...
func SetParams(params ...Param) {
	for _, p := range params {
		p()
	}
}

// ...
const (
	DefaultMaxActiveDelay    Height = 4032
	DefaultActiveDelayFactor Height = 32
)

// https://lbry.io/news/hf1807
const (
	DefaultOriginalClaimExpirationTime       Height = 262974
	DefaultExtendedClaimExpirationTime       Height = 2102400
	DefaultExtendedClaimExpirationForkHeight Height = 278160
)

var (
	paramMaxActiveDelay                    = DefaultMaxActiveDelay
	paramActiveDelayFactor                 = DefaultActiveDelayFactor
	paramOriginalClaimExpirationTime       = DefaultOriginalClaimExpirationTime
	paramExtendedClaimExpirationTime       = DefaultExtendedClaimExpirationTime
	paramExtendedClaimExpirationForkHeight = DefaultExtendedClaimExpirationForkHeight
)

// ResetParams ...
func ResetParams() Param {
	return func() {
		paramMaxActiveDelay = DefaultMaxActiveDelay
		paramActiveDelayFactor = DefaultActiveDelayFactor

		paramOriginalClaimExpirationTime = DefaultOriginalClaimExpirationTime
		paramExtendedClaimExpirationTime = DefaultExtendedClaimExpirationTime
		paramExtendedClaimExpirationForkHeight = DefaultExtendedClaimExpirationForkHeight
	}
}

// MaxActiveDelay ...
func MaxActiveDelay(h Height) Param {
	return func() { paramMaxActiveDelay = h }
}

// ActiveDelayFactor ...
func ActiveDelayFactor(f Height) Param {
	return func() { paramActiveDelayFactor = f }
}

// OriginalClaimExpirationTime ...
func OriginalClaimExpirationTime(h Height) Param {
	return func() { paramOriginalClaimExpirationTime = h }
}

// ExtendedClaimExpirationTime ...
func ExtendedClaimExpirationTime(h Height) Param {
	return func() { paramExtendedClaimExpirationTime = h }
}

// ExtendedClaimExpirationForkHeight ...
func ExtendedClaimExpirationForkHeight(at Height) Param {
	return func() { paramExtendedClaimExpirationForkHeight = at }
}
