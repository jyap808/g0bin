g0bin
=====

g0bin is a client side encrypted pastebin.  The server has zero knowledge of pasted data.  Data is encrypted/decrypted in the browser using 256 bits AES.

g0bin is a Go port of [0bin](https://github.com/sametmax/0bin/) (written in Python).  0bin in turn is an implementation of the [ZeroBin](https://github.com/sebsauvage/ZeroBin/) project (written in PHP).  [These diagrams](http://sebsauvage.net/wiki/doku.php?id=php:zerobin#how_does_it_work) best describe how the encryption and decryption process works.

This project was created mostly as a Go learning exercise through converting a project I use often.  It also serves as a great sample project in Go since it only uses the standard library.

Here are some elements that are have been implemented.

 * [Nested HTML Templates](http://stackoverflow.com/questions/9573644/go-appengine-how-to-structure-templates-for-application/9587616#9587616)
 * [Using Anonymous Structs to pass data to HTML Templates](http://julianyap.com/2013/09/23/using-anonymous-structs-to-pass-data-to-templates-in-golang.html)
 * [Hot configuration reload from a JSON configuration file](http://openmymind.net/Golang-Hot-Configuration-Reload/)
 * [HTTP server logging](https://groups.google.com/forum/#!topic/golang-nuts/s7Xk1q0LSU0)

Install
-------

Clone this repository, build it and run it.

    git clone https://github.com/jyap808/g0bin.git
    cd g0bin
    go build
    ./g0bin

To run g0bin on a different port, modify the Port setting in config.json.

The configuration of g0bin can also be reloaded by sending a HUP signal to the process.

    kill -HUP [PROCESS ID]

Other
-----

This project modifies the Python implementation and cleans things up to make them more generic.  It has the following changes:

 * Remove extra header links.
 * Remove extra link options.
 * Remove extra layout details.
 * Set Burn after reading as default option.

Here are some items which are in the Python implemention which have not been implemented.

 * Paste counter - Display a tiny counter for pastes created
 * Names/links to insert in the menu bar
 * Handling of Max Size
 * Paste ID length
 * Clone paste

Copyright (c) 2013 Julian Yap

[MIT License](https://github.com/jyap808/g0bin/blob/master/LICENSE)

