package randomword

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const adjectives = `accountable
active
agile
anxious
appreciative
balanced
boisterous
bold
boring
brash
brave
bright
busy
chatty
cheerful
chipper
comedic
courteous
daring
delightful
empathic
excellent
exhausted
fantastic
fearless
fluent
forgiving
friendly
funny
generous
grateful
grouchy
grumpy
happy
hilarious
humble
impressive
insightful
intelligent
interested
kind
lean
nice
noisy
optimistic
patient
persistent
proud
quick
quiet
reflective
relaxed
reliable
responsible
responsive
rested
restless
shiny
shy
silly
sleepy
smart
spontaneous
surprised
sweet
talkative
terrific
thankful
timely
tired
turbulent
unexpected
wacky
wise
zany`

const nouns = `aardvark
alligator
antelope
baboon
badger
bandicoot
bat
bear
bilby
bongo
bonobo
buffalo
bushbuck
camel
cassowary
cat
cheetah
chimpanzee
chipmunk
civet
crane
crocodile
dingo
dog
dugong
duiker
echidna
eland
elephant
emu
fossa
fox
gazelle
gecko
gelada
genet
gerenuk
giraffe
gnu
gorilla
grysbok
hartebeest
hedgehog
hippopotamus
hyena
hyrax
impala
jackal
jaguar
kangaroo
klipspringer
koala
kob
kookaburra
kudu
lemur
leopard
lion
lizard
llama
lynx
manatee
mandrill
meerkat
mongoose
mouse
numbat
nyala
okapi
oribi
oryx
ostrich
otter
panda
pangolin
panther
parrot
platypus
porcupine
possum
puku
quokka
quoll
rabbit
ratel
raven
reedbuck
rhinoceros
roan
sable
serval
shark
sitatunga
springhare
squirrel
swan
tasmaniandevil
tiger
topi
toucan
turtle
wallaby
warthog
waterbuck
wildebeest
wolf
wolverine
wombat
zebra`

type Generator struct {
	r *rand.Rand
}

func NewGenerator() Generator {
	return Generator{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (gen Generator) Babble() string {
	return fmt.Sprintf("%s-%s-%s", gen.RandomAdjective(), gen.RandomNoun(), gen.RandomTwoLetters())
}

func (gen Generator) RandomAdjective() string {
	return randomElement(gen.r, adjectives)
}

func (gen Generator) RandomNoun() string {
	return randomElement(gen.r, nouns)
}

func (gen Generator) RandomTwoLetters() string {
	var asciiLetterA = 97
	letterOne := string(rune(gen.r.Intn(26) + asciiLetterA))
	letterTwo := string(rune(gen.r.Intn(26) + asciiLetterA))
	return letterOne + letterTwo
}

func randomElement(r *rand.Rand, fullList string) string {
	wordList := strings.Split(fullList, "\n")
	randomWordIndex := r.Int() % len(wordList)

	return wordList[randomWordIndex]
}
