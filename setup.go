package main

//June 2022, Julius Schoen / R.M. Spicer,  GPL 3 license
//written to show how to move objects smoothly with the g3n game engine

//Gopher model used from gokoban, which gives the following link:
//Gopher 3D model from: https://github.com/StickmanVentures/go-gopher-model
//The model was changed by me (in blender) to get the origins to geometry,
//parented everything to body,
//changed the orientation to be correct (face looking down negative Y axis in blender),
//and added an animation.

import (
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/g3n/engine/animation"
	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/loader/gltf"
	"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/texture"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/util/logger"
	"github.com/g3n/engine/window"
)

//hold game settings
type GameApp struct {
	*app.Application // Embedded standard application object

	Scene *core.Node //master scene

	Camera   *camera.Camera
	orbit    *camera.OrbitControl
	Ship     *core.Node
	Log      *logger.Logger
	DirData  string
	ambLight *light.Ambient
	grid     *helper.Grid
}

//Demo basic struct
type moveGopher struct {
	grid *helper.Grid
	//our main character
	gopher *core.Node
	//not moved in render loop, but tied to quatSlerp LookAt's
	soloGopher *core.Node

	//supporting actors
	vecVelocity, vecVelocityPaused                  math32.Vector3
	vecMovement, vecMovementGoal, vecMovementPaused math32.Vector3
	vecRotation, vecRotationGoal, vecRotationPaused math32.Vector3

	//special vector to demonstrate smooth changes in velocity
	vecAppVelocity, vecAppVelocityGoal math32.Vector3

	//used for slerp'ing
	fromQuat, toQuat math32.Quaternion

	//bit part players
	sphere1, sphere2 *graphic.Mesh
	infoS            *graphic.Sprite
	infoT            *texture.Texture2D
	font             *text.Font
	anims, soloanims []*animation.Animation
	glb, sologlb     *gltf.GLTF
	//this becomes a pointer to INode, this then assigned to soloGopher
	mesh, solomesh core.INode
}

var (

	//allows switching movement between gopher and camera
	currentNode  *core.Node
	nodeIsGopher bool //sad variable needed to deal with differences between objects and cameras

	//used in quaternion slerping through a go routine
	vecLookAt, vecLookAtLooker, vecLookAtTarget math32.Vector3
	rotMatrix                                   math32.Matrix4

	//these are the vectors that makes flying / looking where you're running possible
	vecViewTmp, vecViewForward, vecViewRight, vecViewUp math32.Vector3

	//controls the current movement mode, simple translation vs. flying
	mvType, mvCnt int = 0, 0

	//chooses three objects in order for blue gopher to LookAt
	ToggleLookAtTarget int = -1
)

//save some garbage collection
var (
	usePos       math32.Vector3
	flDifference = float32(0.0)
	vecUpHat     = math32.NewVector3(0, 1, 0)
	//didn't use
	//vecRightHat                        = math32.NewVector3(1, 0, 0)
	//vecScreenHat                       = math32.NewVector3(0, 0, 1)
	incRot, incLinear, incAcceleration = float32(0.0), float32(0.0), float32(2)
)

//"constant" vars
var (
	zeroVector   math32.Vector3 = *math32.NewVector3(0, 0, 0)
	cameraVector math32.Vector3 = *math32.NewVector3(15, 4, -2)
)

const (
	mvTranslate = iota
	mvFly
)

const (
	incrementRotTranslate, incrementRotFly = float32(0.02), float32(0.004)
	incrementLinear                        = float32(0.005)
	progName                               = "Movement demo for g3n"
	execName                               = "g3nmovedemo"
)

// CreateGame and return a pointer to it
func CreateGame() *GameApp {

	game := new(GameApp)

	game.DirData = game.checkDirData("data")

	game.Application = app.App()
	//this is in v0.2.1....
	//game.Application = app.App(1280, 920, "New title")

	game.IWindow.(*window.GlfwWindow).SetSize(1280, 920)
	game.IWindow.(*window.GlfwWindow).SetTitle("g3n move demo, see instructions.txt")
	game.IWindow.(*window.GlfwWindow).SetFullscreen(false)

	game.setupLogs()
	game.setupBasics()
	game.setupShip()

	// Subscribe window to events
	game.Subscribe(window.OnKeyDown, game.onKeyDown)

	game.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) { game.OnWindowResize() })

	// Trigger window resize to recompute UI
	game.OnWindowResize()

	return game
}

//Initialize the moveGopher demo
func (mg *moveGopher) Initialize(gm *GameApp) {

	//load main character
	mg.loadGLTF(0, filepath.Join(gm.DirData, "gopher.glb"))
	gm.Scene.Add(mg.mesh)
	mg.gopher = mg.mesh.GetNode()
	mg.gopher.SetScale(0.3, 0.3, 0.3)
	mg.gopher.SetPosition(0, 0, 0)

	//load our QuatSlerp model, a glb with an animation
	mg.loadGLTF(1, filepath.Join(gm.DirData, "sologopher.glb"))
	gm.Scene.Add(mg.solomesh)
	mg.soloGopher = mg.solomesh.GetNode()
	mg.soloGopher.SetScale(0.6, 0.6, 0.6)
	mg.soloGopher.SetPosition(-5, 4, 3)

	//load geometries
	//texfile := a.DirData + "/images/"
	tex1, err := texture.NewTexture2DFromImage(filepath.Join(gm.DirData, "checkerboard.jpg"))
	if err != nil {
		gm.Log.Fatal("Error loading texture: %s", err)
	}
	tex1.SetWrapS(gls.REPEAT)
	tex1.SetWrapT(gls.REPEAT)
	tex1.SetRepeat(2, 2)
	// Creates sphere 1
	geom1 := geometry.NewSphere(1, 32, 32)
	mat1 := material.NewStandard(&math32.Color{1, 1, 1})
	mat1.AddTexture(tex1)
	mg.sphere1 = graphic.NewMesh(geom1, mat1)
	mg.sphere1.SetPosition(-10, 4, 10)
	gm.Scene.Add(mg.sphere1)

	geom2 := geometry.NewSphere(2, 32, 32)
	mat2 := material.NewStandard(&math32.Color{1, 1, 1})
	mat2.AddTexture(tex1)
	mg.sphere2 = graphic.NewMesh(geom2, mat2)
	mg.sphere2.SetPosition(0, 4, 10)
	gm.Scene.Add(mg.sphere2)

	//load fonts and setup message sprite

	fontfile := filepath.Join(gm.DirData, "FreeSans.ttf")
	font, err := text.NewFont(fontfile)
	if err != nil {
		gm.Log.Fatal(err.Error())
	}
	font.SetLineSpacing(1.0)
	font.SetPointSize(10)
	font.SetDPI(96)

	font.SetFgColor(math32.NewColor4("blue", 1))
	//font.SetBgColor(math32.NewColor4("sandybrown", 0.8))
	font.SetBgColor(math32.NewColor4("yellow", 0.8))

	mg.font = font

	stext := "start"
	//swidth, sheight := font.MeasureText(stext)
	swidth, sheight := 250, 64
	canvas := text.NewCanvas(swidth, sheight, math32.NewColor4("white", 1))
	canvas.DrawText(0, 0, stext, font)
	mg.infoT = texture.NewTexture2DFromRGBA(canvas.RGBA)
	mat := material.NewStandard(math32.NewColor("white"))
	mat.AddTexture(mg.infoT)
	aspect := float32(swidth) / float32(sheight)
	mesh := graphic.NewSprite(aspect, 1, mat)
	//mesh.SetPosition(-2, 1, -1)
	mesh.SetPosition(2, -1, -5)
	mg.infoS = mesh
	gm.Camera.Add(mesh)
	//gm.Scene().Add(mesh)

	// Set background color to gray
	gm.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	currentNode = mg.gopher
	nodeIsGopher = true
	//gm.AddEvent(evergogame.WsEVKeyD, mg.onKeyDown)

	mvType = mvTranslate
	//gm.Camera.Remove(gm.Ship)
	mg.doReset(gm)
}

// check that the data directory exists
func (gm *GameApp) checkDirData(dirDataName string) string {
	if _, err := os.Stat(dirDataName); err != nil {
		panic(err)
	}

	dirData, err := filepath.Abs(dirDataName)
	if err != nil {
		panic(err)
	}
	return dirData
}

// Toggle Full Screen on/off
func (gm *GameApp) ToggleFullScreen() {
	toggle := gm.IWindow.(*window.GlfwWindow).Fullscreen()
	gm.IWindow.(*window.GlfwWindow).SetFullscreen(!toggle)
}

// Quits game
func (gm *GameApp) Quit() {
	gm.Log.Info("Hit Quit func()")
	gm.Exit()
}

// default handler for window resize events.
func (gm *GameApp) OnWindowResize() {
	// Get framebuffer size and set the viewport accordingly
	width, height := gm.GetFramebufferSize()
	gm.Gls().Viewport(0, 0, int32(width), int32(height))
	gm.Camera.SetAspect(float32(width) / float32(height))
}

//attend to logging settings
func (gm *GameApp) setupLogs() {
	gm.Log = logger.New("g3nmovedemo", nil)
	gm.Log.AddWriter(logger.NewConsole(false))
	gm.Log.SetFormat(logger.FTIME | logger.FMICROS)
	gm.Log.SetLevel(logger.INFO)

	gm.Log.Info("%s starting", progName)
	gm.Log.Info("OpenGL version: %s", gm.Gls().GetString(gls.VERSION))
	gm.Log.Info("Using data directory:%s", gm.DirData)
}

//attend to basic game objects, scene, cameras, orbit and basic lights
func (gm *GameApp) setupBasics() {

	// Create main scene
	gm.Scene = core.NewNode() //master holder
	gm.Scene.SetName("master")

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(gm.Scene)

	width, height := gm.GetFramebufferSize()
	gm.Gls().Viewport(0, 0, int32(width), int32(height))
	aspect := float32(width) / float32(height)

	gm.Camera = camera.New(aspect)
	gm.Camera.SetPosition(0, 4, 5)

	gm.Scene.Add(gm.Camera)

	// Show axis helper
	ah := helper.NewAxes(3)
	gm.Scene.Add(ah)

	gm.grid = helper.NewGrid(50, 1, math32.NewColor("darkgray"))
	gm.Scene.Add(gm.grid)

	// Add white ambient light to the scene
	gm.ambLight = light.NewAmbient(math32.NewColor("white"), 0.8)
	gm.Scene.Add(gm.ambLight)

	// Create and add directional white light to the scene
	dirLight := light.NewDirectional(math32.NewColor("white"), 1.0)
	dirLight.SetPosition(1, 0, 0)
	gm.Scene.Add(dirLight)

	// Create orbit control and set limits
	gm.orbit = camera.NewOrbitControl(gm.Camera)
	gm.orbit.SetEnabled(camera.OrbitAll)
	gm.orbit.MaxPolarAngle = 2 * math32.Pi / 3
	gm.orbit.MinDistance = 3
	gm.orbit.MaxDistance = 1200
}

//add flying "ship" indicator
func (gm *GameApp) setupShip() {
	model := filepath.Join(gm.DirData, "/ship")
	shipdec, err := obj.Decode(model+".obj", model+".mtl")
	if err != nil {
		panic(err.Error())
	}
	// Create a new node with all the objects in the decoded file and adds it to the scene
	gm.Ship, err = shipdec.NewGroup()
	if err != nil {
		panic(err.Error())
	}
	//gm.Ship.SetScale(0.8, 0.8, 0.8)
	gm.Ship.SetPosition(0, -1, -4)
	gm.Ship.RotateX(-0.1)
	gm.Ship.RotateY(math.Pi / 2)
	gm.Ship.RotateZ(math.Pi / 6)
	//game.camera.Add(game.Ship) //done elsewhere
}

//gltf loader and animation set up, if any
func (mg *moveGopher) loadGLTF(which int, fpath string) {

	ext := filepath.Ext(fpath)
	var err error

	if strings.ToUpper(ext) != ".GLB" {
		panic("Currently only supporting gltf .glb files, invalid: " + fpath)
	}

	switch which {
	case 0:
		mg.glb, err = gltf.ParseBin(fpath)
	case 1:
		mg.sologlb, err = gltf.ParseBin(fpath)
	}
	if err != nil {
		panic(err)
	}

	defaultSceneIdx := 0
	switch which {
	case 0:
		if mg.glb.Scene != nil {
			defaultSceneIdx = *mg.glb.Scene
		}
		mg.mesh, err = mg.glb.LoadScene(defaultSceneIdx)
	case 1:
		if mg.sologlb.Scene != nil {
			defaultSceneIdx = *mg.sologlb.Scene
		}
		mg.solomesh, err = mg.sologlb.LoadScene(defaultSceneIdx)
	}
	if err != nil {
		panic(err)
	}

	if which == 1 {
		// Create animations
		for i := range mg.sologlb.Animations {
			anim, _ := mg.sologlb.LoadAnimation(i)
			anim.SetLoop(true)
			anim.SetSpeed(anim.Speed() * 10)
			mg.soloanims = append(mg.soloanims, anim)
		}
	}

}

//Re-set the up, right, screen normals
func (mv *moveGopher) reset3DNormals() {
	//These "constants" can be changed when they are used, re-set them
	//for now I only use the vecUpHat....
	//vecRightHat.Set(1, 0, 0)
	vecUpHat.Set(0, 1, 0)
	//vecScreenHat.Set(0, 0, 1)
}
