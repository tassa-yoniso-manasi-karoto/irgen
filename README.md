# irgen
Incremental reading card generator for Anki supporting Wikipedia & local HTML files 

This is a CLI tool only. At its core it just splits the HTML file using the heading tags into many notes.
However it is capable of creating exceptionally high context notes that would otherwise be too costly timewise to make by means of gathering individually for each note-to-be the surrounding pictures & tables of the document according to user-defined configuration.

The binary will parse a config.json file located in the directory where it is.

<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/Osteo Forearm.webp">
<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/Wikipedia–Geography of Asia.webp">

### TODO
- add doc for config.json
- move PrefForHiRes to extractors
- "pinned" user-defined group of high-value image that can be displayed none
- ankiconnect support

### About
I originally started this project many, many years ago. It lived first as a shell script, then as python script, then as a very poorly written Go codebase. I am providing it here after a near complete rewrite for public interest.

