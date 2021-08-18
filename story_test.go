package goink

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultStory(t *testing.T) {
	story := Default()
	assert.Equal(t, story, story.current.Story())

	ctx := story.save()
	assert.Equal(t, "start", ctx.Current)

	sec, err := story.Resume(&ctx)
	assert.Nil(t, err)
	assert.Equal(t, "", sec.Text)
	assert.Equal(t, 2, len(sec.Tags))
}

func TestBasicParse(t *testing.T) {
	input := `
	This is a basic parsing test. # TAG_A
	Story will read these lines one by one, # tag b
	And connect them togather... # tag c // comment
	-> END
	`
	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	sec, err := story.Resume(ctx)
	assert.Nil(t, err)

	assert.Equal(t, true, sec.End)
	assert.Equal(t, "end", ctx.Current)
	assert.Equal(t, 5, len(sec.Tags)) // 3 + start_tag + end_tag
}

func TestStoryLoad(t *testing.T) {
	input := `
	This is a basic parsing test. # TAG_A
	Story will read these lines one by one, # tag b
	And connect them togather... # tag c // comment
	-> END
	`
	story := Default()
	err := story.Parse(input)
	story.SetID("ABC")
	assert.Nil(t, err)

	ctx := NewContext()
	ctx.Current = "invalid path"
	_, err = story.Resume(ctx)
	assert.Contains(t, err.Error(), "is not existed")

	_, err = story.Pick(ctx, 0)
	assert.Contains(t, err.Error(), "is not existed")

	ctx.Current = "start"
	_, err = story.Pick(ctx, 0)
	assert.Contains(t, err.Error(), "is not an option")

	ctx = NewContext()
	ctx.Vars["start__i"] = "invalid vars"
	_, err = story.Resume(ctx)
	assert.Contains(t, err.Error(), "is not type of int")
}

func TestInvalidNextNode(t *testing.T) {
	input := `
	This is a basic parsing test. # TAG_A
	Story will read these lines one by one, # tag b
	And connect them togather... # tag c // comment
	== Knot_A
	-> END
	`
	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	errs := story.PostParsing()
	assert.Contains(t, errs[0].Error(), "can not go next")

	input = `
	* opt a
	* opt b
	* opt c
	`
	story = Default()
	err = story.Parse(input)
	assert.Nil(t, err)

	errs = story.PostParsing()
	assert.Contains(t, errs[0].Error(), "can not go next")

	input = `
	* opt a
	* opt b
	* opt c
	- gather -> end
	`
	story = Default()
	err = story.Parse(input)
	assert.Nil(t, err)

	errs = story.PostParsing()
	assert.Nil(t, errs)
}

func TestStandardStory(t *testing.T) {

	fmt.Println("TESTING STANDARD STORY")

	//Todo: how do we get/set variables from the game state (e.g. 'do we have this item' 'have we completed this quest' etc.)
	//Todo: how do we assign quests? where are quests defined?

	input := `
VAR quest_complete_Deliver_The_Egg = true

LONDON, 1872
Residence of Monsieur Phileas Fogg.
-> london

=== london ===
Monsieur Phileas Fogg returned home early from the Reform Club, and in a new-fangled steam-carriage, besides!  
"Passepartout," said he. "We are going around the world!"

+ We have a simple weave here
    And this simple weave has one followup {quest_complete_Deliver_The_Egg}
    ++ And a sub-choice of course
        Which itself has a nested response 
    ++ And a second sub-choice
+ {quest_complete_Deliver_The_Egg} And again, here

- Regardless, we march onward!

+ "Around the world, Monsieur?"
    I was utterly astonished. 
    -> astonished
+ [Nod curtly.] -> nod


=== astonished ===
"You are in jest!" I told him in dignified affront. "You make mock of me, Monsieur."
"I am quite serious."

+ "But of course"
    -> ending


=== nod ===
I nodded curtly, not believing a word of it.
-> ending


=== ending
"We shall circumnavigate the globe within eighty days." He was quite calm as he proposed this wild scheme. "We leave for Paris on the 8:25. In an hour."
+ {not astonished} I was not astonished (SHOULD NOT APPEAR)!
+ {astonished} I was astonished!
+ {does_not_exist} This option should never be shown (SHOULD NOT APPEAR)
+ {not does_not_exist} I did not go a place that didn't exist
+ I was neither
- Agreed, grasshopper.
-> END`

	story := Default()
	err := story.Parse(input)
	assert.Nil(t, err)

	ctx := NewContext()
	ctx.Vars = story.CurrentVars()
	story.PostParsing()
	sec, err := story.Resume(ctx)
	if err != nil {
		fmt.Println(err)
	}

	boolVal := ctx.Vars["quest_complete_Deliver_The_Egg"].(bool)
	assert.True(t, boolVal)

	for {
		//fmt.Println(ctx.Vars)
		fmt.Println(sec.Text)

		if len(sec.Opts) == 0 {
			fmt.Println("STORY OVER")
			break
		}

		for _, choice := range sec.Opts {
			fmt.Println("CHOICE: ", choice)
		}

		sec, err = story.Pick(ctx, 0)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println(ctx.Vars)
}

func BenchmarkBasicStoryParsing(b *testing.B) {
	input := `
	This is a basic parsing test. # TAG_A
	Story will read these lines one by one, # tag b
	And connect them togather... # tag c // comment
	-> END
	`
	for i := 0; i < b.N; i++ {
		story := Default()
		if err := story.Parse(input); err != nil {
			panic(err)
		}
	}
}

func BenchmarkComplexStoryParsing(b *testing.B) {
	input := `
    Hello
	-> Knot
	== Knot
	this is a knot content.
	* {knot > 0} Opt A
	  opt a content -> Knot
	* Opt B -> knot
	* Opt C
	- (gather) gather -> END
	== Knot_B
	this is a knot content.
	* {knot > 0} Opt A
	  opt a content -> Knot
	* Opt B -> knot
	* Opt C
	- (gather) gather -> END
	== Knot_C
	this is a knot content.
	* {knot > 0} Opt A
	  opt a content -> Knot
	* Opt B -> knot
	* Opt C
	- (gather) gather -> END
	`
	for i := 0; i < b.N; i++ {
		story := Default()
		if err := story.Parse(input); err != nil {
			b.Log(err)
		}
	}
}
