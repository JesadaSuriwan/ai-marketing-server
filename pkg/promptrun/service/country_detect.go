package service

import "strings"

// tldCountry maps a ccTLD (lowercase, no leading dot — including the last
// label of compound ccTLDs like "co.th", which is just "th") to its ISO
// 3166-1 alpha-2 country code. Most ccTLDs equal their ISO code; "uk" is the
// one common exception (the UK's ISO code is GB, not UK).
var tldCountry = map[string]string{
	"af": "AF", "al": "AL", "dz": "DZ", "ad": "AD", "ao": "AO", "ag": "AG",
	"ar": "AR", "am": "AM", "au": "AU", "at": "AT", "az": "AZ", "bs": "BS",
	"bh": "BH", "bd": "BD", "bb": "BB", "by": "BY", "be": "BE", "bz": "BZ",
	"bj": "BJ", "bt": "BT", "bo": "BO", "ba": "BA", "bw": "BW", "br": "BR",
	"bn": "BN", "bg": "BG", "bf": "BF", "bi": "BI", "cv": "CV", "kh": "KH",
	"cm": "CM", "ca": "CA", "cf": "CF", "td": "TD", "cl": "CL", "cn": "CN",
	"co": "CO", "km": "KM", "cg": "CG", "cd": "CD", "cr": "CR", "ci": "CI",
	"hr": "HR", "cu": "CU", "cy": "CY", "cz": "CZ", "dk": "DK", "dj": "DJ",
	"dm": "DM", "do": "DO", "ec": "EC", "eg": "EG", "sv": "SV", "gq": "GQ",
	"er": "ER", "ee": "EE", "sz": "SZ", "et": "ET", "fj": "FJ", "fi": "FI",
	"fr": "FR", "ga": "GA", "gm": "GM", "ge": "GE", "de": "DE", "gh": "GH",
	"gr": "GR", "gd": "GD", "gt": "GT", "gn": "GN", "gw": "GW", "gy": "GY",
	"ht": "HT", "hn": "HN", "hk": "HK", "hu": "HU", "is": "IS", "in": "IN",
	"id": "ID", "ir": "IR", "iq": "IQ", "ie": "IE", "il": "IL", "it": "IT",
	"jm": "JM", "jp": "JP", "jo": "JO", "kz": "KZ", "ke": "KE", "ki": "KI",
	"kw": "KW", "kg": "KG", "la": "LA", "lv": "LV", "lb": "LB", "ls": "LS",
	"lr": "LR", "ly": "LY", "li": "LI", "lt": "LT", "lu": "LU", "mg": "MG",
	"mw": "MW", "my": "MY", "mv": "MV", "ml": "ML", "mt": "MT", "mh": "MH",
	"mr": "MR", "mu": "MU", "mx": "MX", "fm": "FM", "md": "MD", "mc": "MC",
	"mn": "MN", "me": "ME", "ma": "MA", "mz": "MZ", "mm": "MM", "na": "NA",
	"nr": "NR", "np": "NP", "nl": "NL", "nz": "NZ", "ni": "NI", "ne": "NE",
	"ng": "NG", "kp": "KP", "mk": "MK", "no": "NO", "om": "OM", "pk": "PK",
	"pw": "PW", "ps": "PS", "pa": "PA", "pg": "PG", "py": "PY", "pe": "PE",
	"ph": "PH", "pl": "PL", "pt": "PT", "qa": "QA", "ro": "RO", "ru": "RU",
	"rw": "RW", "kn": "KN", "lc": "LC", "vc": "VC", "ws": "WS", "sm": "SM",
	"st": "ST", "sa": "SA", "sn": "SN", "rs": "RS", "sc": "SC", "sl": "SL",
	"sg": "SG", "sk": "SK", "si": "SI", "sb": "SB", "so": "SO", "za": "ZA",
	"kr": "KR", "ss": "SS", "es": "ES", "lk": "LK", "sd": "SD", "sr": "SR",
	"se": "SE", "ch": "CH", "sy": "SY", "tw": "TW", "tj": "TJ", "tz": "TZ",
	"th": "TH", "tl": "TL", "tg": "TG", "to": "TO", "tt": "TT", "tn": "TN",
	"tr": "TR", "tm": "TM", "tv": "TV", "ug": "UG", "ua": "UA", "ae": "AE",
	"uk": "GB", "us": "US", "uy": "UY", "uz": "UZ", "vu": "VU", "va": "VA",
	"ve": "VE", "vn": "VN", "ye": "YE", "zm": "ZM", "zw": "ZW",
}

// genericTLDs are technically valid ccTLDs (or gTLDs) that are overwhelmingly
// used for their English-word meaning rather than as a geographic signal —
// treating them as country hits would produce confident-looking false
// positives (an "ai" startup's .ai domain isn't targeting Anguilla). These
// fall through to the LLM classifier instead, same as true generic TLDs.
var genericTLDs = map[string]bool{
	"com": true, "org": true, "net": true, "info": true, "biz": true,
	"edu": true, "gov": true, "mil": true, "int": true, "name": true,
	"io": true, "co": true, "ai": true, "me": true, "tv": true, "to": true,
	"fm": true, "ly": true, "gg": true, "cc": true, "sh": true, "app": true,
	"dev": true, "xyz": true, "shop": true, "store": true, "online": true,
	"site": true, "tech": true, "cloud": true,
}

// detectCountryFromTLD returns an ISO 3166-1 alpha-2 code deterministically
// derived from rawURL's TLD (just the last label — correctly resolves
// compound ccTLDs like "co.th" or "co.uk" too, since "th"/"uk" is still the
// last label there), or "" if the TLD carries no reliable country signal
// (generic/hijacked TLD, or the host couldn't be parsed at all).
func detectCountryFromTLD(rawURL string) string {
	host := normalizeDomain(rawURL)
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	last := parts[len(parts)-1]
	if genericTLDs[last] {
		return ""
	}
	return tldCountry[last]
}

// isValidCountryCode sanity-checks the extraction model's target_country
// guess before it's trusted — two uppercase letters is enough to keep
// obvious model slop ("N/A", "unknown", "global") out of the column without
// requiring an exact match against tldCountry's necessarily-incomplete list.
func isValidCountryCode(code string) bool {
	if len(code) != 2 {
		return false
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
