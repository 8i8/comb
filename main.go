// package comb generates a UUID with 73bits of cryptographically random data
// in its first 10 bytes, and 6 bytes of timestamp data after that, the
// timestamp has a 10th of a millisecond precision and covers a temporal range
// of 892 years before wrapping.  7 bits are used to set values so as to
// remain rfc4122 compatible, comprising of the variant and version
// information, variant future and version 6.
package comb

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/google/uuid"
)

const pkg = "comb"

// NullUUID mimics the behaviour of the sql.Null* types
type NullUUID struct {
	uuid.UUID
	Valid bool
}

func uint64ToBytes(b []byte, n int, v uint64) {
	_ = b[n-1] // early bounds check
	for i := 0; i < n; i++ {
		b[i] = byte(v >> (1 << (n - 1 - i)))
	}
}

func bytesToUint64(b []byte, n int) uint64 {
	lenInt := 8
	_ = b[n-1]           // early bounds check
	if len(b) > lenInt { // 8 bytes maximum.
		panic("byte slice is larger than uint64")
	}
	var val uint64
	for i := 0; i < n; i++ {
		val |= uint64(b[n-1-i]) << (i * lenInt)
	}
	return val
}

// ReadTimeStamp reads the time stamp that is set into a TimeStampedUUID.
func ReadTimeStamp(id uuid.UUID) uint64 {
	return ReadCustomTimeStamp(id, 6)
}

// ReadCustomTimeStamp reads n bytes from the least significant bit of
// the uuid and returns the value contained there as an integer.
func ReadCustomTimeStamp(id uuid.UUID, nBytes int) uint64 {
	return bytesToUint64(id[len(id)-nBytes:], nBytes)
}

// NewTimeStampedUUID returns a UUID with 73bits of cryptographically
// random data in its first 10 bytes, and 6 bytes of timestamp data
// after that, the timestamp has a 10th of a millisecond precision and
// covers a temporal range of 892 years before wrapping.  7 bits are
// used to set values so as to remain rfc4122 compatible, comprising of
// the variant and version information, variant future and version 6.
func NewTimeStampedUUID() (uuid.UUID, error) {
	now, _, err := uuid.GetTime()
	if err != nil {
		return uuid.Nil, fmt.Errorf("NewTimeStampedUUID: %w", err)
	}
	return CustomTimeStampedUUID(rand.Reader, 6, now, time.Millisecond/10, true)
}

func SetTimeStamp(id uuid.UUID, nBytes int, t uuid.Time, res time.Duration) (uuid.UUID, error) {
	// Translate duration into parts per second, the time is already being
	// returned from GetTime in 100th's of a nano second, dividing 1e8 by
	// the resolution gives us the correct numerator when using this time
	// format.
	res = time.Second / 10 / res // Translate duration into parts per second.
	if nBytes > len(id) {
		return id, errors.New("to many bytes to format")
	}

	// Write the first 6 bytes with the least significant 6 bytes of the
	// current Time as measured in 100s of microseconds since 15 Oct 1582.
	mask := uint64(1<<uint64(nBytes*8) - 1)
	timeBytes := uint64(math.Round((float64(t) / float64(res)))) & mask
	uint64ToBytes(id[len(id)-nBytes:], nBytes, timeBytes)
	return id, nil
}

// CustomTimeStampedUUID generates a uuid.UUID with n bytes of time stamp set
// to the given time resolution and the remaining bytes random data.
func CustomTimeStampedUUID(r io.Reader, nBytes int, t uuid.Time, res time.Duration, rfc4122 bool) (uuid.UUID, error) {
	var id uuid.UUID
	const fname = "CustomTimeStampedUUID"
	var err error
	fail := func(err error) (uuid.UUID, error) {
		return id, fmt.Errorf("%s: %w", fname, err)
	}

	id, err = SetTimeStamp(id, nBytes, t, res)
	if err != nil {
		return fail(err)
	}

	// Fill the remaining bytes with values from the io.Reader.
	_, err = io.ReadFull(r, id[:len(id)-nBytes])
	if err != nil {
		return fail(err)
	}

	if rfc4122 {
		// In accordance with rfc4122 Set version to 6, an as yet unspecifed
		// version.
		id[6] = (id[6] & 0x0f) | 0x60 // Version 6
		id[8] = (id[8] & 0x3f) | 0xe0 // Variant is 111, future
	}

	return id, nil
}

// timeRange displays information about the time range available if a
// specific time duration is set to be the length of time represented by
// an integer for the specified word size.
func timeRange(wordSize uint64, timeResolution time.Duration) {
	const avgYear = 365.24219
	const secPerDay = 88400

	partsPerSecond := time.Second / timeResolution // Translate duration into parts per second.
	normalised := 1 / float64(partsPerSecond)
	units := int64(math.Pow(2, float64(wordSize))) // Total units available.

	seconds := units / int64(partsPerSecond) // Length of that time in seconds.
	unitsRmn := units % int64(partsPerSecond)

	days := seconds / secPerDay // Length of that time in days.
	secondsRmn := seconds % secPerDay

	years := int64(float64(days) / avgYear) //Length of that time in years.
	daysRemainder := (float64(days) / avgYear) - float64(years)
	daysRmn := int64(daysRemainder * avgYear)

	fmt.Printf("%d years %d days %f seconds\n", years, daysRmn, float64(secondsRmn)+(normalised*float64(unitsRmn)))
}
