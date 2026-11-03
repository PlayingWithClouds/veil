package library

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
)

// defaultCategory is one category of the list a new library starts with:
// its name, the other spellings sites tag it with (the name itself always
// matches) and its subcategories.
type defaultCategory struct {
	name     string
	terms    []string
	children []defaultCategory
}

// leaf is a category without subcategories.
func leaf(name string, terms ...string) defaultCategory {
	return defaultCategory{name: name, terms: terms}
}

// group is a top-level category made of its subcategories.
func group(name string, children ...defaultCategory) defaultCategory {
	return defaultCategory{name: name, children: children}
}

// defaultCategories is the taxonomy a new library starts with, grouped, each
// with the spellings sites commonly use. The user edits, adds and deletes
// from there.
var defaultCategories = []defaultCategory{
	group("Body",
		leaf("Big tits", "big boobs", "busty", "huge tits", "huge boobs", "big breasts", "big natural tits", "big natural boobs"),
		leaf("Small tits", "small boobs", "tiny tits", "flat chested", "petite tits"),
		leaf("Natural tits", "natural boobs", "natural breasts"),
		leaf("Fake tits", "fake boobs", "implants", "silicone tits", "enhanced"),
		leaf("Puffy nipples", "big nipples", "long nipples"),
		leaf("Big ass", "big butt", "bubble butt", "phat ass", "pawg", "big booty"),
		leaf("Curvy", "thick", "voluptuous", "thicc", "hourglass"),
		leaf("BBW", "chubby", "plump", "fat", "ssbbw"),
		leaf("Petite", "tiny", "small", "spinner"),
		leaf("Skinny", "slim", "thin", "slender"),
		leaf("Athletic", "fit", "muscular", "fitness", "abs"),
		leaf("Tall", "amazon", "long legs"),
		leaf("Pregnant", "preggo"),
		leaf("Hairy", "hairy pussy", "bush", "unshaved", "natural bush"),
		leaf("Shaved", "shaved pussy", "bald pussy", "smooth pussy"),
		leaf("Tattoos", "tattoo", "tattooed", "inked"),
		leaf("Piercings", "piercing", "pierced", "pierced nipples", "nipple piercing"),
		leaf("Big dick", "big cock", "huge cock", "huge dick", "monster cock", "bbc"),
		leaf("Feet", "foot fetish", "soles", "toes"),
	),
	group("Hair",
		leaf("Blonde", "blond", "blondes"),
		leaf("Brunette", "brunettes", "brown hair"),
		leaf("Redhead", "ginger", "red hair", "redheads"),
		leaf("Black hair", "black haired"),
		leaf("Colored hair", "pink hair", "blue hair", "purple hair", "dyed hair", "green hair"),
		leaf("Short hair", "pixie cut", "bob"),
		leaf("Pigtails", "braids", "ponytail", "twintails"),
	),
	group("Age",
		leaf("Teen (18+)", "teen", "teens", "18", "19", "18 year old", "barely legal"),
		leaf("College", "coed", "student", "university"),
		leaf("MILF", "milfs", "hot mom", "yummy mummy"),
		leaf("Mature", "cougar", "older woman", "matures"),
		leaf("GILF", "granny", "grandma", "old lady"),
		leaf("Age gap", "old young", "older younger", "old and young", "older man"),
	),
	group("Ethnicity",
		leaf("Asian", "asians", "oriental"),
		leaf("Japanese", "jav", "japan"),
		leaf("Chinese", "china"),
		leaf("Korean", "korea"),
		leaf("Thai", "thailand"),
		leaf("Filipina", "pinay", "philippines"),
		leaf("Indian", "desi", "india"),
		leaf("Arab", "arabic", "middle eastern", "persian", "turkish"),
		leaf("Ebony", "black", "black women", "african"),
		leaf("Latina", "latin", "latinas", "hispanic", "mexican", "colombian"),
		leaf("Brazilian", "brazil"),
		leaf("Russian", "russia", "ukrainian"),
		leaf("Czech", "czech republic"),
		leaf("European", "euro", "europe"),
		leaf("British", "uk", "english"),
		leaf("German", "germany", "deutsch"),
		leaf("French", "france"),
		leaf("Italian", "italy"),
		leaf("Scandinavian", "swedish", "danish", "norwegian", "nordic"),
		leaf("Interracial", "ir", "mixed race"),
	),
	group("Acts",
		leaf("Blowjob", "bj", "blow job", "oral", "sucking", "cock sucking"),
		leaf("Deepthroat", "deep throat", "throat fuck", "throatpie"),
		leaf("Handjob", "hand job", "hj"),
		leaf("Titjob", "titty fuck", "tit fuck", "boobjob", "tittyfuck"),
		leaf("Footjob", "foot job"),
		leaf("Pussy licking", "cunnilingus", "eating pussy", "pussy eating"),
		leaf("Rimming", "rimjob", "rim job", "ass licking", "ass eating"),
		leaf("Anal", "anal sex", "ass fuck", "assfuck", "butt fuck", "first anal"),
		leaf("Double penetration", "dp", "dvp", "dap", "double anal", "double vaginal"),
		leaf("Creampie", "internal cumshot", "cum inside", "anal creampie"),
		leaf("Facial", "facials", "cum on face"),
		leaf("Cumshot", "cum shot", "cumshots", "cum"),
		leaf("Swallow", "cum swallow", "cum in mouth", "swallowing"),
		leaf("Squirting", "squirt", "female ejaculation"),
		leaf("Masturbation", "masturbating", "fingering", "jerking off", "jerk off"),
		leaf("Toys", "toy", "dildo", "vibrator", "sex toy", "sex toys", "magic wand", "hitachi", "buttplug", "butt plug"),
		leaf("Fisting", "fist", "anal fisting"),
		leaf("Gaping", "gape", "gapes"),
		leaf("69", "sixty nine", "sixtynine"),
		leaf("Doggystyle", "doggy", "doggy style"),
		leaf("Cowgirl", "riding", "reverse cowgirl"),
		leaf("Missionary"),
		leaf("Kissing", "making out", "french kissing"),
		leaf("Orgasm", "orgasms", "real orgasm", "female orgasm"),
		leaf("Edging", "ruined orgasm", "tease and denial"),
		leaf("Facesitting", "face sitting", "queening", "smothering"),
		leaf("Scissoring", "tribbing", "tribadism"),
	),
	group("People",
		leaf("Solo", "solo female", "solo girl"),
		leaf("Solo male", "male masturbation"),
		leaf("Couple", "couples"),
		leaf("Lesbian", "lesbians", "girl on girl", "girl girl"),
		leaf("Gay", "gay male", "gay sex"),
		leaf("Bisexual", "bi", "bisexual male", "mmf bisexual"),
		leaf("Trans", "transgender", "shemale", "ts", "tgirl", "ladyboy", "transsexual", "tranny"),
		leaf("Threesome", "ffm", "mmf", "3some"),
		leaf("Foursome", "4some", "swap"),
		leaf("Group sex", "orgy", "group", "sex party"),
		leaf("Gangbang", "gang bang", "reverse gangbang"),
		leaf("Bukkake"),
		leaf("Amateur", "homemade", "real couple", "amateurs"),
		leaf("Pornstar", "pornstars", "porn star"),
		leaf("Cuckold", "cuck", "hotwife", "cuckquean"),
		leaf("Swingers", "swinger", "wife swap", "swapping"),
		leaf("Cheating", "cheat", "unfaithful"),
	),
	group("Fetish & kink",
		leaf("BDSM", "s&m", "sm", "sadomasochism"),
		leaf("Bondage", "tied up", "rope", "shibari", "restrained", "bound"),
		leaf("Femdom", "female domination", "dominatrix", "mistress", "domination"),
		leaf("Maledom", "male domination", "master"),
		leaf("Submissive", "sub", "slave", "obedience"),
		leaf("Spanking", "spank", "paddling", "caning"),
		leaf("Rough sex", "rough", "hard fuck", "hardcore rough"),
		leaf("Choking", "breath play"),
		leaf("Humiliation", "degradation", "verbal humiliation"),
		leaf("Pegging", "strapon", "strap on", "strap-on"),
		leaf("CFNM", "clothed female nude male"),
		leaf("CBT", "ball busting", "ballbusting", "cock and ball torture"),
		leaf("Chastity", "chastity cage", "locked"),
		leaf("Latex", "pvc", "rubber", "catsuit"),
		leaf("Leather"),
		leaf("Nylons", "nylon", "pantyhose", "tights"),
		leaf("Foot worship", "feet worship", "toe sucking", "foot licking"),
		leaf("Lactation", "lactating", "breast milk", "milking"),
		leaf("Pissing", "piss", "pee", "peeing", "golden shower", "watersports"),
		leaf("Spit", "spitting", "drool", "sloppy"),
		leaf("Smoking", "smoking fetish", "cigarette"),
		leaf("Tickling", "tickle"),
		leaf("Wax play", "hot wax", "candle wax"),
		leaf("Electro", "estim", "e-stim", "electrostim"),
		leaf("Sissy", "sissification", "crossdressing", "crossdresser", "feminization"),
		leaf("JOI", "jerk off instruction", "jerk off instructions", "cei"),
		leaf("Findom", "financial domination", "paypig"),
		leaf("Voyeur", "voyeurism", "hidden cam", "spy cam", "peeping"),
		leaf("Exhibitionism", "exhibitionist", "flashing", "public nudity"),
		leaf("Upskirt", "downblouse", "nip slip"),
		leaf("Roleplay", "role play", "fantasy", "taboo"),
		leaf("Step family", "stepmom", "stepsister", "step sister", "step mom", "stepdaughter", "stepbrother", "step brother", "family roleplay", "fauxcest"),
		leaf("Medical", "doctor", "gyno", "exam", "clinic"),
		leaf("Hypnosis", "hypno", "mind control", "trance"),
		leaf("Giantess", "size difference", "shrinking", "vore"),
		leaf("Inflation", "belly inflation", "balloon", "looner"),
		leaf("Mud & food", "wet and messy", "wam", "food play", "messy"),
		leaf("Sploshing", "pie", "slime"),
		leaf("Armpits", "armpit", "hairy armpits", "armpit licking"),
		leaf("Belly", "belly button", "navel"),
		leaf("Muscle worship", "female muscle", "fbb", "bodybuilder"),
		leaf("Sweaty", "sweat", "after workout"),
	),
	group("Exotic",
		leaf("Hentai", "anime", "manga", "ecchi"),
		leaf("3D & CGI", "3d", "cgi", "animation", "animated", "sfm", "blender"),
		leaf("Cartoon", "toon", "comic", "parody"),
		leaf("Futanari", "futa", "dickgirl"),
		leaf("Monster", "tentacle", "tentacles", "monster girl", "alien", "orc"),
		leaf("Furry", "yiff", "anthro"),
		leaf("Cosplay", "costume", "cosplayer"),
		leaf("Fantasy creatures", "elf", "succubus", "demon girl", "vampire", "fairy"),
		leaf("Robot & sci-fi", "robot", "android", "cyborg", "sci-fi"),
		leaf("AI generated", "ai", "ai art", "ai porn", "deepfake"),
		leaf("Sex machine", "fucking machine", "sex machines", "machine"),
		leaf("Sex doll", "sex dolls", "love doll", "realdoll"),
		leaf("Glory hole", "gloryhole", "glory holes"),
		leaf("Clown & circus", "clown", "circus"),
		leaf("Body paint", "bodypaint", "body painting"),
		leaf("Oil", "oiled", "oily", "baby oil", "lotion"),
		leaf("Pregnant fantasy", "breeding", "impregnation"),
	),
	group("Clothing",
		leaf("Lingerie", "underwear", "bra", "panties", "thong"),
		leaf("Stockings", "garter", "garter belt", "thigh highs", "fishnets", "fishnet"),
		leaf("High heels", "heels", "stilettos", "boots"),
		leaf("Bikini", "swimsuit", "swimwear", "one piece"),
		leaf("Yoga pants", "leggings", "spandex", "yoga"),
		leaf("Jeans", "denim", "shorts"),
		leaf("Skirt", "miniskirt", "mini skirt"),
		leaf("Uniform", "uniforms", "police", "military", "cheerleader", "stewardess"),
		leaf("Nurse"),
		leaf("Maid", "french maid", "housekeeper"),
		leaf("Office", "secretary", "business suit", "boss"),
		leaf("Glasses", "nerd", "geek"),
		leaf("Corset", "bustier"),
		leaf("Socks", "knee socks", "knee highs"),
		leaf("Dress", "evening dress", "sundress"),
		leaf("Wedding", "bride", "wedding dress"),
		leaf("Sportswear", "sports bra", "tennis", "gym clothes"),
		leaf("Christmas", "santa", "xmas", "halloween", "holiday"),
	),
	group("Setting",
		leaf("Outdoor", "outdoors", "outside", "nature", "forest", "garden"),
		leaf("Public", "public sex", "in public", "risky public"),
		leaf("Beach", "nude beach", "seaside"),
		leaf("Pool", "swimming pool", "poolside", "jacuzzi", "hot tub"),
		leaf("Shower", "bath", "bathroom", "bathtub", "wet"),
		leaf("Bedroom", "bed"),
		leaf("Kitchen"),
		leaf("Car", "car sex", "backseat", "taxi"),
		leaf("Massage", "massage parlor", "masseuse", "nuru", "happy ending"),
		leaf("Gym", "workout", "fitness room"),
		leaf("Hotel", "hotel room", "motel"),
		leaf("Casting", "audition", "fake casting", "first time"),
		leaf("Party", "club", "college party", "bachelor party"),
		leaf("Vacation", "holiday trip", "travel", "resort"),
		leaf("Camping", "tent", "camper"),
		leaf("Webcam", "cam", "cam girl", "camgirl", "live cam"),
		leaf("Backstage", "behind the scenes", "bts", "bloopers"),
		leaf("Dorm", "dorm room", "frat"),
		leaf("Sauna", "spa", "steam room"),
		leaf("Pickup", "public agent", "picked up", "stranger"),
	),
	group("Style",
		leaf("POV", "point of view"),
		leaf("Close-up", "closeup", "close up", "pussy closeup"),
		leaf("Romantic", "sensual", "passionate", "lovemaking", "erotic"),
		leaf("Softcore", "soft core", "teasing", "tease"),
		leaf("Hardcore", "hard core"),
		leaf("Artistic", "art", "art nude", "fine art", "nude art"),
		leaf("Glamour", "glamor", "posing", "photoshoot", "photo shoot"),
		leaf("Striptease", "strip", "stripping", "stripper", "lap dance", "pole dance"),
		leaf("Dancing", "dance", "twerk", "twerking"),
		leaf("Vintage", "retro", "classic", "70s", "80s", "90s"),
		leaf("Compilation", "compilations", "pmv", "hmv", "cumpilation"),
		leaf("VR", "virtual reality", "vr porn", "180", "360"),
		leaf("Vertical", "tiktok", "onlyfans", "selfie"),
		leaf("Funny", "fails", "comedy"),
		leaf("ASMR", "whisper", "whispering", "soft spoken"),
		leaf("Dirty talk", "talking dirty", "talk"),
		leaf("Slow motion", "slow mo", "slowmo"),
	),
}

// SeedDefaultCategories gives a library that never had categories the
// default list. A library that has (or had: trashed ones count) categories
// is left as the user made it; AddDefaultCategories brings defaults back.
func (l *Library) SeedDefaultCategories(ctx context.Context) error {
	existing, err := l.database.Int(ctx, `SELECT COUNT(*) FROM "Category"`, nil)
	if err != nil {
		return err
	}
	if existing > 0 {
		return nil
	}
	_, err = l.AddDefaultCategories(ctx)
	return err
}

// AddDefaultCategories adds the default categories the library lacks (by
// slug), under their default parents, and returns how many it added. A
// default whose parent is missing goes to the top.
func (l *Library) AddDefaultCategories(ctx context.Context) (int, error) {
	return l.addDefaults(ctx, defaultCategories, "")
}

// addDefaults adds the categories of list (and their subcategories) that
// are missing, under parentID ("" = top level).
func (l *Library) addDefaults(ctx context.Context, list []defaultCategory, parentID string) (int, error) {
	added := 0
	for position, entry := range list {
		id, created, err := l.defaultCategoryID(ctx, entry, parentID, position)
		if err != nil {
			return added, fmt.Errorf("default category %s: %w", entry.name, err)
		}
		if created {
			added++
		}
		children, err := l.addDefaults(ctx, entry.children, id)
		added += children
		if err != nil {
			return added, err
		}
	}
	return added, nil
}

// defaultCategoryID returns the id of the default's category, creating it
// when no category has its slug (a trashed one counts: the slug stays
// taken); created says whether it did.
func (l *Library) defaultCategoryID(ctx context.Context, entry defaultCategory, parentID string, position int) (string, bool, error) {
	existing, err := l.database.Strings(ctx, `SELECT "id" FROM "Category" WHERE "slug" = $slug`,
		db.Vars{"slug": slugOf(entry.name)})
	if err != nil {
		return "", false, err
	}
	if len(existing) > 0 {
		return existing[0], false, nil
	}
	name := entry.name
	edit := CategoryEdit{Name: &name, MatchTerms: termsOf(entry), Position: &position}
	if parentID != "" {
		edit.ParentID = &parentID
	}
	created, err := l.CreateCategory(ctx, edit)
	if err != nil {
		return "", false, err
	}
	return created.Id, true, nil
}

// termsOf is what a default matches: a category of its own its name and
// spellings; a group only through its subcategories, so a tag named "Body"
// isn't all of them.
func termsOf(entry defaultCategory) *[]string {
	terms := []string{}
	if len(entry.children) == 0 {
		terms = append([]string{entry.name}, entry.terms...)
	}
	return &terms
}
