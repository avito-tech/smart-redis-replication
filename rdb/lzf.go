package rdb

// Taken from Golly: https://github.com/tav/golly/blob/master/lzf/lzf.go
// Removed part that gets outputLength from data
// nolint:gocyclo
func lzfDecompress(input []byte, outputLength uint32) (output []byte) {
	inputLength := uint32(len(input))

	var backref int64
	var ctrl, iidx, length, oidx uint32

	output = make([]byte, outputLength)
	iidx = 0

	for iidx < inputLength {
		// Get the control byte.
		ctrl = uint32(input[iidx])
		iidx++

		if ctrl < (1 << 5) {
			// The control byte indicates a literal reference.
			ctrl++
			if oidx+ctrl > outputLength {
				return nil
			}

			// Safety check.
			if iidx+ctrl > inputLength {
				return nil
			}

			for {
				output[oidx] = input[iidx]
				iidx++
				oidx++
				ctrl--
				if ctrl == 0 {
					break
				}
			}
		} else {
			// The control byte indicates a back reference.
			length = ctrl >> 5
			backref = int64(oidx - ((ctrl & 31) << 8) - 1)

			// Safety check.
			if iidx >= inputLength {
				return nil
			}

			// It's an extended back reference. Read the extended length before
			// reading the full back reference location.
			if length == 7 {
				length += uint32(input[iidx])
				iidx++
				// Safety check.
				if iidx >= inputLength {
					return nil
				}
			}

			// Put together the full back reference location.
			backref -= int64(input[iidx])
			iidx++

			if oidx+length+2 > outputLength {
				return nil
			}

			if backref < 0 {
				return nil
			}

			output[oidx] = output[backref]
			oidx++
			backref++
			output[oidx] = output[backref]
			oidx++
			backref++
			for {
				output[oidx] = output[backref]
				oidx++
				backref++
				length--
				if length == 0 {
					break
				}
			}

		}
	}
	return output
}
