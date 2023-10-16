package convertors

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/rotisserie/eris"
)

// GenerateRandomName generates random name for `run`.
func GenerateRandomName() (string, error) {
	p, err := rand.Int(rand.Reader, big.NewInt(int64(len(PREDICATES))))
	if err != nil {
		return "", eris.Wrap(err, "error getting random integer number")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(NOUNS))))
	if err != nil {
		return "", eris.Wrap(err, "error getting random integer number")
	}
	i, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return "", eris.Wrap(err, "error getting random integer number")
	}
	return fmt.Sprintf("%s-%s-%d", PREDICATES[p.Int64()], NOUNS[n.Int64()], i), nil
}

var (
	NOUNS = []string{
		"ant",
		"ape",
		"asp",
		"auk",
		"bass",
		"bat",
		"bear",
		"bee",
		"bird",
		"boar",
		"bug",
		"calf",
		"carp",
		"cat",
		"chimp",
		"cod",
		"colt",
		"conch",
		"cow",
		"crab",
		"crane",
		"croc",
		"crow",
		"cub",
		"deer",
		"doe",
		"dog",
		"dolphin",
		"donkey",
		"dove",
		"duck",
		"eel",
		"elk",
		"fawn",
		"finch",
		"fish",
		"flea",
		"fly",
		"foal",
		"fowl",
		"fox",
		"frog",
		"gnat",
		"gnu",
		"goat",
		"goose",
		"grouse",
		"grub",
		"gull",
		"hare",
		"hawk",
		"hen",
		"hog",
		"horse",
		"hound",
		"jay",
		"kit",
		"kite",
		"koi",
		"lamb",
		"lark",
		"loon",
		"lynx",
		"mare",
		"midge",
		"mink",
		"mole",
		"moose",
		"moth",
		"mouse",
		"mule",
		"newt",
		"owl",
		"ox",
		"panda",
		"penguin",
		"perch",
		"pig",
		"pug",
		"quail",
		"ram",
		"rat",
		"ray",
		"robin",
		"roo",
		"rook",
		"seal",
		"shad",
		"shark",
		"sheep",
		"shoat",
		"shrew",
		"shrike",
		"shrimp",
		"skink",
		"skunk",
		"sloth",
		"slug",
		"smelt",
		"snail",
		"snake",
		"snipe",
		"sow",
		"sponge",
		"squid",
		"squirrel",
		"stag",
		"steed",
		"stoat",
		"stork",
		"swan",
		"tern",
		"toad",
		"trout",
		"turtle",
		"vole",
		"wasp",
		"whale",
		"wolf",
		"worm",
		"wren",
		"yak",
		"zebra",
	}

	PREDICATES = []string{
		"abundant",
		"able",
		"abrasive",
		"adorable",
		"adaptable",
		"adventurous",
		"aged",
		"agreeable",
		"ambitious",
		"amazing",
		"amusing",
		"angry",
		"auspicious",
		"awesome",
		"bald",
		"beautiful",
		"bemused",
		"bedecked",
		"big",
		"bittersweet",
		"blushing",
		"bold",
		"bouncy",
		"brawny",
		"bright",
		"burly",
		"bustling",
		"calm",
		"capable",
		"carefree",
		"capricious",
		"caring",
		"casual",
		"charming",
		"chill",
		"classy",
		"clean",
		"clumsy",
		"colorful",
		"crawling",
		"dapper",
		"debonair",
		"dashing",
		"defiant",
		"delicate",
		"delightful",
		"dazzling",
		"efficient",
		"enchanting",
		"entertaining",
		"enthused",
		"exultant",
		"fearless",
		"flawless",
		"fortunate",
		"fun",
		"funny",
		"gaudy",
		"gentle",
		"gifted",
		"glamorous",
		"grandiose",
		"gregarious",
		"handsome",
		"hilarious",
		"honorable",
		"illustrious",
		"incongruous",
		"indecisive",
		"industrious",
		"intelligent",
		"inquisitive",
		"intrigued",
		"invincible",
		"judicious",
		"kindly",
		"languid",
		"learned",
		"legendary",
		"likeable",
		"loud",
		"luminous",
		"luxuriant",
		"lyrical",
		"magnificent",
		"marvelous",
		"masked",
		"melodic",
		"merciful",
		"mercurial",
		"monumental",
		"mysterious",
		"nebulous",
		"nervous",
		"nimble",
		"nosy",
		"omniscient",
		"orderly",
		"overjoyed",
		"peaceful",
		"painted",
		"persistent",
		"placid",
		"polite",
		"popular",
		"powerful",
		"puzzled",
		"rambunctious",
		"rare",
		"rebellious",
		"respected",
		"resilient",
		"righteous",
		"receptive",
		"redolent",
		"resilient",
		"rogue",
		"rumbling",
		"salty",
		"sassy",
		"secretive",
		"selective",
		"sedate",
		"serious",
		"shivering",
		"skillful",
		"sincere",
		"skittish",
		"silent",
		"smiling",
		"sneaky",
		"sophisticated",
		"spiffy",
		"stately",
		"suave",
		"stylish",
		"tasteful",
		"thoughtful",
		"thundering",
		"traveling",
		"treasured",
		"trusting",
		"unequaled",
		"upset",
		"unique",
		"unleashed",
		"useful",
		"upbeat",
		"unruly",
		"valuable",
		"vaunted",
		"victorious",
		"welcoming",
		"whimsical",
		"wistful",
		"wise",
		"worried",
		"youthful",
		"zealous",
	}
)
