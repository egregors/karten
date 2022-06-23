# karten

Words memorizing app with CLI 

* Deutsch
  * add words
    * with auto translation, by data provider (verbformen.com)
    * manually
  * learn words

---
<div align="center">

[![Build Status](https://github.com/egregors/karten/actions/workflows/go.yml/badge.svg)](https://github.com/egregors/karten/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/egregors/karten)](https://goreportcard.com/report/github.com/egregors/karten)

</div>

> **Warning**
> App in an active development now, so I bet you gonna lose you date a few times, until first stable release ;)

## Install

Clone this repo `git clone git@github.com:egregors/karten.git`, run `make install`.

## Usage

Just run `karten` to exercise, or `karten -a` to add new words.

| short | long  | description                                      |
|-------|-------|--------------------------------------------------|
| -a    | --add | Add new words into your dictionary               |
|       | --dbg | Debug mode to print some additional information. |

### Add new words
[//]: # (Add asciinem with adding process: meta & manual)

When you try to add a new word, Karten will try to get some info (translation, forms, grammar) about this word by 
particular data provider. If it fails, you can add your own translation for the word.

### Learn words
[//]: # (Add asciinem with learning process)

All words have his own rating, how good you know them. It's stars from 0 to 5. When 
you remember the word, one star will be added. Otherwise – removed. Besides, words lose 
his rating during the time.

Each time you are run the program, Karten will choose 20 words with the smallest rating 
for you.

## Contributing

Bug reports, bug fixes and new features are always welcome.
Please open issues and submit pull requests for any new code.