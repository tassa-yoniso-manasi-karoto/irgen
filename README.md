# irgen
Incremental reading card generator supporting Wikipedia &amp; local files.

This is a CLI tool only. At its core it just split the HTML file using the heading tags into many notes.
However it is capable of creating exceptionally high context notes that would otherwise be too costly timewise to make by mean of gathering for each and every future note the nearby picture & table of the document according to user-defined configuration.

<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/Osteo Forearm.webp">
<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/Wikipedia–Geography of Asia.webp">

### TODO
- PUSH CARD CSS
- ankiconnect support
- test windows build
- add doc for config.json

### About
I originally started this project many, many years ago. It lived first as a shell script, then as python script, then as a very poorly written Go codebase. I am providing it here after a near complete rewrite for public interest.
