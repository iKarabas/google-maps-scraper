package gmaps

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	olc "github.com/google/open-location-code/go"
)

func ParseSearchResults(raw []byte) ([]*Entry, error) {
	var data []any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	container, ok := data[0].([]any)
	if !ok || len(container) == 0 {
		return nil, fmt.Errorf("invalid business list structure")
	}

	items := getNthElementAndCast[[]any](container, 1)
	if len(items) < 2 {
		return nil, fmt.Errorf("empty business list")
	}

	entries := make([]*Entry, 0, len(items)-1)

	for i := 1; i < len(items); i++ {
		arr, ok := items[i].([]any)
		if !ok {
			continue
		}

		business := getNthElementAndCast[[]any](arr, 14)

		var entry Entry

		entry.ID = getNthElementAndCast[string](business, 0)
		entry.Title = getNthElementAndCast[string](business, 11)
		entry.Categories = toStringSlice(getNthElementAndCast[[]any](business, 13))
		if len(entry.Categories) > 0 {
			entry.Category = entry.Categories[0]
		}

		entry.WebSite = getNthElementAndCast[string](business, 7, 0)

		entry.ReviewRating = getNthElementAndCast[float64](business, 4, 7)
		entry.ReviewCount = int(getNthElementAndCast[float64](business, 4, 8))
		entry.ReviewsLink = getNthElementAndCast[string](business, 4, 3, 0)


		fullAddress := getNthElementAndCast[[]any](business, 2)

		entry.CompleteAddress = Address{
			Borough:    getNthElementAndCast[string](business, 183, 1, 0),
			Street:     getNthElementAndCast[string](business, 183, 1, 1), 
			City:       getNthElementAndCast[string](business, 183, 1, 3),
			PostalCode: getNthElementAndCast[string](business, 183, 1, 4),
			State:      getNthElementAndCast[string](business, 183, 1, 5),
			Country:    getNthElementAndCast[string](business, 183, 1, 6),
		}

		entry.Address = func() string {
			sb := strings.Builder{}

			for i, part := range fullAddress {
				if i > 0 {
					sb.WriteString(", ")
				}

				sb.WriteString(fmt.Sprintf("%v", part))
			}

			return sb.String()
		}()

		entry.Latitude = getNthElementAndCast[float64](business, 9, 2)
		entry.Longtitude = getNthElementAndCast[float64](business, 9, 3)
		entry.Phone = strings.ReplaceAll(getNthElementAndCast[string](business, 178, 0, 0), " ", "")
		entry.OpenHours = getHours(business)
		entry.Status = getNthElementAndCast[string](business, 34, 4, 4)
		entry.Timezone = getNthElementAndCast[string](business, 30)
		entry.DataID = getNthElementAndCast[string](business, 10)

		// Extract CID from DataID (convert hex to decimal)
		if entry.DataID != "" {
			parts := strings.Split(entry.DataID, ":")
			if len(parts) == 2 {
				// Remove "0x" prefix from the second part and convert hex to decimal
				hexPart := strings.TrimPrefix(parts[1], "0x")
				if cid, err := strconv.ParseUint(hexPart, 16, 64); err == nil {
					entry.Cid = fmt.Sprintf("%d", cid)
				}
			}
		}


		// Add About parsing
		aboutI := getNthElementAndCast[[]any](business, 100, 1)

		for j := range aboutI {
			el := getNthElementAndCast[[]any](aboutI, j)
			about := About{
				ID:   getNthElementAndCast[string](el, 0),
				Name: getNthElementAndCast[string](el, 1),
			}

			optsI := getNthElementAndCast[[]any](el, 2)

			for k := range optsI {
				opt := Option{
					Enabled: (getNthElementAndCast[float64](optsI, k, 2, 1, 0, 0)) == 1,
					Name:    getNthElementAndCast[string](optsI, k, 1),
				}

				if opt.Name != "" {
					about.Options = append(about.Options, opt)
				}
			}

			entry.About = append(entry.About, about)
		}

		entry.PlusCode = olc.Encode(entry.Latitude, entry.Longtitude, 10)

		entries = append(entries, &entry)
	}

	return entries, nil
}

func toStringSlice(arr []any) []string {
	ans := make([]string, 0, len(arr))
	for _, v := range arr {
		ans = append(ans, fmt.Sprintf("%v", v))
	}

	return ans
}
