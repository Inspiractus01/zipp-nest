package main

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultNestPort = 9090

// 256 short words — index = byte value (0-255)
var wordList = [256]string{
	"ace", "age", "aid", "aim", "air", "ale", "ant", "arc",
	"arm", "art", "ash", "axe", "bay", "bed", "bee", "big",
	"bit", "box", "boy", "bud", "bug", "bun", "bus", "cab",
	"can", "cap", "car", "cat", "cow", "cup", "cut", "dad",
	"dam", "day", "den", "dew", "dig", "dim", "dip", "dog",
	"dot", "dry", "ear", "eat", "egg", "elf", "elk", "elm",
	"end", "era", "eye", "far", "fat", "fig", "fin", "fit",
	"fly", "fog", "fox", "fun", "fur", "gap", "gas", "gem",
	"gin", "gum", "gun", "gut", "gym", "ham", "hat", "hay",
	"hen", "hip", "hog", "hop", "hot", "hub", "hug", "hum",
	"ice", "ill", "imp", "ink", "inn", "ion", "ivy", "jab",
	"jam", "jar", "jaw", "jet", "jig", "job", "jot", "joy",
	"jug", "keg", "key", "kid", "kin", "kit", "lab", "lag",
	"lap", "law", "lay", "led", "leg", "lid", "lip", "lit",
	"log", "lot", "low", "lug", "mad", "map", "mat", "mob",
	"mop", "mud", "mug", "nap", "net", "nil", "nip", "nod",
	"nor", "nun", "oak", "oar", "odd", "oil", "old", "orb",
	"ore", "owl", "own", "pad", "pan", "paw", "pay", "pea",
	"peg", "pen", "pet", "pie", "pig", "pin", "pit", "pod",
	"pop", "pot", "pub", "pug", "pun", "pup", "ram", "ran",
	"rat", "raw", "ray", "red", "rid", "rig", "rim", "rip",
	"rob", "rod", "rot", "row", "rub", "rug", "rum", "run",
	"rut", "rye", "sad", "sap", "sat", "saw", "say", "sea",
	"set", "sew", "shy", "sin", "sip", "sir", "sit", "ski",
	"sky", "sly", "sob", "son", "sow", "spa", "spy", "sub",
	"sue", "sum", "sun", "tab", "tan", "tap", "tar", "tax",
	"tea", "tie", "tin", "tip", "toe", "ton", "top", "tow",
	"tug", "tun", "two", "urn", "van", "vat", "vet", "via",
	"vim", "vow", "wag", "war", "wax", "web", "wed", "wet",
	"wig", "win", "wit", "woe", "won", "wry", "yak", "yam",
	"yap", "yew", "zap", "zen", "zit", "cod", "cob", "cog",
	"cot", "cue", "cud", "dab", "dud", "dun", "fad", "foe",
}

// encodeNestCode encodes an IPv4 address into 4 words separated by dashes.
// Example: "100.71.53.15" → "kin-hay-fig-big"
func encodeNestCode(ip string) (string, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid IP")
	}
	words := make([]string, 4)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return "", fmt.Errorf("invalid IP")
		}
		words[i] = wordList[n]
	}
	return strings.Join(words, "-"), nil
}
