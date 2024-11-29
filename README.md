***Project status: alpha***. 

Incremental reading note generator for Anki supporting Wikipedia & local HTML files

At its core it just splits the HTML file using the heading tags into many notes.

However it is capable of creating exceptionally high context notes that would otherwise be too costly timewise to make by means of gathering individually for each note-to-be the surrounding pictures & tables of the document according to user-defined configuration.


<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/osteo_forearm.webp">
<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/wiki_geo_asia.webp">

## How to use the GUI
Make sure to have Anki open with AnkiConnect installed. *(If the video below doesn't play open it in a new tab)*

<img width=600 src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/GUI.webp">

## How to use
### Prerequisites
IRGen requires the Anki addon **[AnkiConnect](https://ankiweb.net/shared/info/2055492159)** to communicate with Anki, otherwise it will output a tab-separated file containing the notes to be imported manually.

.docx / .odt / .epub can't be processed directly but you can save them as HTML using your favorite word processor, which can in turn be passed to irgen.

**You must edit manually your note's list of fields, front, back templates and CSS first.** You need to create a "RealTitle" and "Context" field. Your fields should be as follows:

<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/fields.png">

**Copy paste the content of files from the [note](https://github.com/tassa-yoniso-manasi-karoto/irgen/tree/main/note) templates to their corresponding template field of the IR3 Notetype in Anki.**

Due to Anki or Qt shenanigans, copying the exact same template I have on my IR3 Notetype to the IR3 Notetype of a new profile gave me a different webrender. I have had to change it but it mostly look the same as in my clips.

For the CLI:
- The binary will parse a config.json file for default value located in the directory where it itself is located.
- Before running irgen you need to provide the path to the directory where your media files are located in the config.json. Irgen will automatically download or copy images from the document to this location.
This directory is named "collection.media". Please see the [anki docs](https://docs.ankiweb.net/files.html?highlight=collection.medi#file-locations) for how to find it.

On Windows you may want to install Notepad++ to edit that JSON file. Please be aware the syntax of JSON requires that the "\\" in the path to your collection.media directory to be escaped using another "\\" as shown here:

<img width=700 src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/config_windows.png">

### Running irgen from the CLI (windows)

On Windows open the folder where irgen is located using the file manager and click on "File" (upper left corner), "Open Windows PowerShell".

A command line interface will open where you can now pass an URL or the path to a local HTML file like this:

<img src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/powershell.png">

## config.json
Taking this local HTML file as reference, I will explain the entries of config.json. Let's take as reference for my examples the note-to-be located under "least important" title and that contains Lorem ipsum with the picture of a snake:

<img width=500 src="https://github.com/tassa-yoniso-manasi-karoto/irgen/blob/main/demo/example.webp">

_(h1, h2... are the heading tags that will appear in the raw HTML you don't need to add them to the text, this is just to illustrate)_

**Briefly, these are the keys of the config.json and how they will shape the output :**
- **CollectionMedia** : the path to your "collection.media" folder
- **DestDir** : you can optionally set a default destination directory for the .txt file
- **MaxTitles** : this is the max number of headings that will appear in Anki. With it set to 3, the card of my example will get a RealTitle like this "quite important: less important: least important", omitting "Very important title".
- **Functions** are each paired with a scope. The scope is the relative position above the heading of a note-to-be. In the example above: 1 would correspond to the heading "less important" located one level above in importance to the heading of the note that contains Lorem ipsum with the snake. Currently only FromSuperior and FromSuperiorAndDescendants are implemented. ***FromSuperior*** will retrieve only the image of the phylogenetic tree where as ***FromSuperiorAndDescendants*** will capture both the image of the phylogenetic tree and the one with the frog.
- **ResXMax** and **ResYMax**: on wikipedia each image is available in various resolutions and irgen will automatically download the highest quality available but you can limit the maximal resolution accepted using these values.

## Download
**See [releases](https://github.com/tassa-yoniso-manasi-karoto/irgen/releases/).**

On Linux you will need **gtk3** and **webkit** installed.

On Windows you will be prompted to install Microsoft's Edge-based Webview2 on launch.

Apple makes it a pain to cross-compile programs with a GUI for MacOS so only the CLI binary is available. ¯\\\_(ツ)\_/¯

## TODO
- reimplement MvAddendumToCxt
- "pinned" user-defined group of high-value image that can be displayed none

## About
I originally started this project many, many years ago. It lived first as a shell script, then as python script, then as a very poorly written Go codebase. I am providing it here after a near complete rewrite for public interest. (edit: the code is still pretty bad tbh)



