g0bin
=====

g0bin is a client side encrypted pastebin.  The server has zero knowledge of pasted data.  Data is encrypted/decrypted in the browser using 256 bits AES.

**DEMO: http://g0bin-demo.appspot.com**

g0bin is a Go port of [0bin](https://github.com/sametmax/0bin/) (written in Python).  0bin in turn is an implementation of the [ZeroBin](https://github.com/sebsauvage/ZeroBin/) project (written in PHP).

This project was created mostly as a Go learning exercise through converting a project I use often.  It also serves as a great sample project in Go since it only uses the standard library.

Here are some elements that are have been implemented.

 * [Nested HTML Templates](http://stackoverflow.com/questions/9573644/go-appengine-how-to-structure-templates-for-application/9587616#9587616)
 * [Using Anonymous Structs to pass data to HTML Templates](http://julianyap.com/2013/09/23/using-anonymous-structs-to-pass-data-to-templates-in-golang.html)
 * [Hot configuration reload from a JSON configuration file](http://openmymind.net/Golang-Hot-Configuration-Reload/)
 * [HTTP server logging](https://groups.google.com/forum/#!topic/golang-nuts/s7Xk1q0LSU0)

NOTE: The demo was modified to run on Google App Engine using the Datastore API.

How it works
------------

When pasting a text into g0bin:

![Encryption image](http://julianyap.com/g0bin/images/encryption.png)

 * You paste your text in the browser and click the “Submit” button.
 * A random 256 bits key is generated in the browser.
 * Data is compressed and encrypted with AES using Javascript libraries.
 * Encrypted data is sent to the server and stored.
 * The browser displays the final URL with the key.
 * The key is never transmitted to the server, which therefore cannot decrypt data.

When opening a g0bin URL:

![Decryption image](http://julianyap.com/g0bin/images/decryption.png)

 * The browser requests encrypted data from the server
 * The decryption key is in the anchor part of the URL (#…) **which is never sent to the server**.
 * Data is decrypted in the browser using the key and displayed.

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

