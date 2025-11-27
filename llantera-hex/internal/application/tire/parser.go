package tireapp

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

type measurementData struct {
	width        int
	profile      *int
	rim          float64
	construction string
	tubeType     string
	plyRating    string
	loadIndex    string
	speedIndex   string
	remainder    string
}

var (
	metricRegex    = regexp.MustCompile(`(?i)^(\d{3})\s*/\s*(\d{2})\s*([R-])\s*(\d{2})(.*)$`)
	motoRegex      = regexp.MustCompile(`(?i)^(\d{2,3})\s*/\s*(\d{2,3})\s*-\s*(\d{2})(.*)$`)
	flotationRegex = regexp.MustCompile(`(?i)^(\d{2,3})\s*X\s*(\d{1,2}\.\d{1,2})\s*([R-])\s*(\d{2})(.*)$`)
	agriRegex      = regexp.MustCompile(`(?i)^(\d{1,2}\.\d)\s*([R-])\s*(\d{2})(.*)$`)
	loadSpeedRegex = regexp.MustCompile(`(?i)(\d{2,3})([A-Z]{1,2})`)
	plyRegex       = regexp.MustCompile(`(?i)(\d{1,2})\s*PR`)
)

var brandDictionary = map[string]string{
	"AB":          "AB Tires",
	"AURORA":      "Aurora Tires",
	"BS":          "Bridgestone",
	"BRIDGESTONE": "Bridgestone",
	"DAYTON":      "Dayton",
	"DOUBLE COIN": "Double Coin",
	"FS":          "Firestone",
	"FIRESTONE":   "Firestone",
	"FUZION":      "Fuzion",
	"GDY":         "Goodyear",
	"GOODYEAR":    "Goodyear",
	"GOO":         "Goodride",
	"HAN":         "Hankook",
	"HANKOOK":     "Hankook",
	"KUM":         "Kumho",
	"KUMHO":       "Kumho",
	"LAUFENN":     "Laufenn",
	"OTR":         "OTR Tires",
	"OTRAS":       "Otras Marcas",
	"PIRELLI":     "Pirelli",
	"SUM":         "Sumitomo",
	"SUMITOMO":    "Sumitomo",
	"TOR":         "Tornel",
	"TORNEL":      "Tornel",
}

var typeDictionary = map[string]string{
	"PS":                     "Pasajero",
	"PASAJERO":               "Pasajero",
	"PASAJERO RADIAL":        "Pasajero Radial (PSR)",
	"PSR":                    "Pasajero Radial (PSR)",
	"LT":                     "Camioneta Convencional",
	"LTS":                    "Light Truck Convencional (LTS)",
	"LTR":                    "Light Truck Radial (LTR)",
	"ST":                     "Special Trailer (ST)",
	"TBR":                    "Truck & Bus Radial (TBR)",
	"LT R":                   "Light Truck Radial (LTR)",
	"LTA":                    "Light Truck Radial (LTR)",
	"IND":                    "Industrial Radial",
	"INDUSTRIAL":             "Industrial Radial",
	"MOTO CONVENCIONAL":      "Moto Convencional",
	"MOTO RADIAL":            "Moto Radial",
	"AGR":                    "Agrícola Radial",
	"AGRICOLA":               "Agrícola Radial",
	"CAMION RADIAL":          "Camión Radial",
	"CAMION CONVENCIONAL":    "Camión Convencional",
	"CAMIONETA RADIAL":       "Camioneta Radial",
	"CAMIONETA CONVENCIONAL": "Camioneta Convencional",
	"LLANTA TEMPORAL":        "Llanta Temporal",
}

func parseMeasurement(raw string) measurementData {
	cleaned := strings.TrimSpace(strings.ToUpper(raw))
	data := measurementData{remainder: cleaned}

	switch {
	case metricRegex.MatchString(cleaned):
		parts := metricRegex.FindStringSubmatch(cleaned)
		data.width = atoi(parts[1])
		prof := atoi(parts[2])
		data.profile = &prof
		data.construction = mapConstruction(parts[3])
		data.rim = atof(parts[4])
		data.remainder = strings.TrimSpace(parts[5])
	case flotationRegex.MatchString(cleaned):
		parts := flotationRegex.FindStringSubmatch(cleaned)
		data.width = atoi(parts[1])
		prof := int(math.Round(atof(parts[2]) * 25.4))
		data.profile = &prof
		data.construction = mapConstruction(parts[3])
		data.rim = atof(parts[4])
		data.remainder = strings.TrimSpace(parts[5])
	case motoRegex.MatchString(cleaned):
		parts := motoRegex.FindStringSubmatch(cleaned)
		data.width = atoi(parts[1])
		prof := atoi(parts[2])
		data.profile = &prof
		data.construction = "D"
		data.rim = atof(parts[3])
		data.remainder = strings.TrimSpace(parts[4])
	case agriRegex.MatchString(cleaned):
		parts := agriRegex.FindStringSubmatch(cleaned)
		data.width = int(math.Round(atof(parts[1]) * 25.4))
		data.profile = nil
		data.construction = mapConstruction(parts[2])
		data.rim = atof(parts[3])
		data.remainder = strings.TrimSpace(parts[4])
	}

	if ls := loadSpeedRegex.FindStringSubmatch(cleaned); len(ls) == 3 {
		data.loadIndex = ls[1]
		data.speedIndex = ls[2]
	}

	if ply := plyRegex.FindStringSubmatch(cleaned); len(ply) == 2 {
		data.plyRating = strings.ToUpper(strings.TrimSpace(ply[0]))
	}

	if strings.Contains(cleaned, "TL") {
		data.tubeType = "TL"
	} else if strings.Contains(cleaned, "TT") {
		data.tubeType = "TT"
	}

	return data
}

func normalizeBrand(alias, fallback string) string {
	key := strings.ToUpper(strings.TrimSpace(alias))
	if key == "" {
		key = strings.ToUpper(strings.TrimSpace(fallback))
	}
	if val, ok := brandDictionary[key]; ok {
		return val
	}
	if key == "" {
		return "Otras Marcas"
	}
	return strings.Title(strings.ToLower(key))
}

func normalizeType(abbr, description string) string {
	candidates := []string{abbr, description}
	for _, candidate := range candidates {
		key := strings.ToUpper(strings.TrimSpace(candidate))
		if key == "" {
			continue
		}
		if val, ok := typeDictionary[key]; ok {
			return val
		}
	}
	return "Otros"
}

func mapConstruction(token string) string {
	if strings.EqualFold(token, "R") {
		return "R"
	}
	if strings.EqualFold(token, "D") {
		return "D"
	}
	return ""
}

func atoi(val string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(val))
	return v
}

func atof(val string) float64 {
	clean := strings.ReplaceAll(val, ",", ".")
	clean = strings.TrimSpace(clean)
	f, _ := strconv.ParseFloat(clean, 64)
	return f
}

func parsePrice(raw string) float64 {
	cleaned := strings.ReplaceAll(raw, "$", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" || cleaned == "-" {
		return 0
	}
	v, _ := strconv.ParseFloat(cleaned, 64)
	return v
}
