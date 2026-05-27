package customer

import "testing"

func TestFormatAccountNumberZeroPadsToSeedFormat(t *testing.T) {
	// Account numbers are minted server-side in the seed format ISP-#### (see
	// the seeder, which starts at ISP-1001). The formatter owns that rule.
	cases := []struct {
		seq  int
		want string
	}{
		{1, "ISP-0001"},
		{1001, "ISP-1001"},
		{1234, "ISP-1234"},
	}
	for _, tc := range cases {
		if got := FormatAccountNumber(tc.seq); got != tc.want {
			t.Errorf("FormatAccountNumber(%d) = %q, want %q", tc.seq, got, tc.want)
		}
	}
}
