package names

import "math/rand/v2"

var adjectives = []string{
	"amber", "bold", "brave", "bright", "calm", "clever", "cosmic", "crisp", "curious", "daring",
	"eager", "fancy", "fierce", "gentle", "golden", "happy", "hidden", "humble", "jolly", "keen",
	"kind", "lively", "lucky", "merry", "misty", "noble", "polite", "proud", "quick", "quiet",
	"rapid", "rusty", "shiny", "silent", "silver", "sleepy", "smooth", "snowy", "solid", "steady",
	"sunny", "swift", "tidy", "tiny", "vivid", "warm", "wild", "wise", "witty", "zesty",
}

var nouns = []string{
	"badger", "beaver", "bison", "cedar", "comet", "coral", "crane", "delta", "dolphin", "eagle",
	"falcon", "fern", "finch", "fox", "glacier", "harbor", "hawk", "heron", "lake", "lemur",
	"lynx", "maple", "meadow", "moose", "nebula", "otter", "owl", "panda", "pebble", "pine",
	"puffin", "quartz", "raven", "reef", "river", "robin", "salmon", "sparrow", "spruce", "stone",
	"storm", "tiger", "trout", "tundra", "valley", "walrus", "willow", "wolf", "wren", "yak",
}

func Random() string {
	return adjectives[rand.IntN(len(adjectives))] + "-" + nouns[rand.IntN(len(nouns))]
}
