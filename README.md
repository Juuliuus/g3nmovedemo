
# g3nmovedemo - a simple demo on how to smoothly move objects in g3n

The purpose of this demo is to show how to achieve smooth motion while
moving objects or the camera, in both simple translation mode and a
"flying" mode.

It can also serve as a playground for testing ideas before committing
them to your game code.

I use non-standard keys (ie., no wasd) so that it is obvious what axis
one is manipulating.

For this reason the first run through, at least, should be done in
accompaniment with the instructions.txt file.

# About g3n

[G3N engine](https://github.com/g3n/engine) is a 3D game engine written in Go. 

Also see [G3N](https://github.com/g3n) for related links.

# Dependencies for installation

g3nmovedemo only depends on the G3N game engine and so has the same
dependencies as the engine. See those dependencies at the link above.


# Installation

In order to run, build, and or install you will need Go installed on
your system. Search for "golang download**, the process is quite
simple.

Either clone/fork this repo to a folder of your choice, or, from the
code button on this page, select to download a zip file, unzip that in
a the folder of your choice.

From that folder use either "go run ." to run a temporary copy, or "go
build ." to build an executable in that folder.


When you do either of these Go will download the g3n engine and
anything it depends on, if you don't already have it on your system. 

The first run or build will may take a little longer while if things
must be downloaded.


Be sure to have a copy of the instructions.txt file open so you can
walk through the available commands and features.

# Regarding the gopher model

Gopher model was derived from the same model used in [gokoban](https://github.com/danaugrs/gokoban), which
gives the following link: 

Gopher 3D model from:
https://github.com/StickmanVentures/go-gopher-model

For the purposes of this demo the model was changed by me (in Blender)
to get the origins to geometry, parented everything to body, changed
the orientation to be correct (face looking down negative Y axis), and
added an animation.


