package twitch

// paginate drives a Relay-style cursor loop. page runs one query starting after
// the given cursor, asking for first items; it appends the records it decodes to
// the caller's slice and returns how many it added, the next cursor, and whether
// the connection has more pages. paginate calls it until limit is reached, a
// page returns nothing, or there are no more pages. limit <= 0 means "page
// through everything the API will give".
func paginate(limit int, page func(after string, first int) (added int, next string, more bool, err error)) error {
	total := 0
	after := ""
	for {
		first := defaultPageSize
		if limit > 0 {
			remaining := limit - total
			if remaining <= 0 {
				return nil
			}
			if remaining < first {
				first = remaining
			}
		}
		added, next, more, err := page(after, first)
		if err != nil {
			return err
		}
		total += added
		if added == 0 || !more || next == "" {
			return nil
		}
		if limit > 0 && total >= limit {
			return nil
		}
		after = next
	}
}
