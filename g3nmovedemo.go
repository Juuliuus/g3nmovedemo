package main

//June 2022, Julius Schoen / R.M. Spicer,  GPL 3 license
//written to show how to move objects smoothly with the g3n game engine

import (
	"fmt"
	"time"

	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/text"
	"github.com/g3n/engine/window"
)

//mundane stuff in setup.go, this file concentrates on the movement routines
//and so is a mix of game and moveGopher methods

var demo *moveGopher

func main() {
	game := CreateGame()

	demo = &moveGopher{}
	demo.Initialize(game)

	game.Application.Run(game.Update)
}

// Game's render loop
func (gm *GameApp) Update(rend *renderer.Renderer, deltaTime time.Duration) {

	dtime := float32(deltaTime.Seconds())

	gm.Gls().Clear(gls.COLOR_BUFFER_BIT | gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT)

	// Render scene
	demo.Update(dtime)
	rend.Render(gm.Scene, gm.Camera)

	gui.Manager().TimerManager.ProcessTimers()

}

//moeGopher Render loop
func (mg *moveGopher) Update(dtime float32) {

	// Label routines
	canvas := text.NewCanvas(250, 64, math32.NewColor4("white", 1))
	canvas.DrawText(0, 0, mg.getCurrentInfo(), mg.font)
	mg.infoT.SetFromRGBA(canvas.RGBA)

	//This is the linear demo in translate mode that moves sphere1 around
	mg.vecAppVelocity.SetZ(Approach(mg.vecAppVelocityGoal.Z, mg.vecAppVelocity.Z, dtime))
	usePos = mg.sphere1.Position() //sadly can't work with Position() directly...
	mg.sphere1.SetPositionVec(usePos.Add(&mg.vecAppVelocity))

	switch mvType {

	case mvTranslate:
		usePos = currentNode.Position()
		currentNode.SetPositionVec(usePos.Add(&mg.vecVelocity))
		currentNode.RotateX(mg.vecRotation.X)
		currentNode.RotateY(mg.vecRotation.Y)
		currentNode.RotateZ(mg.vecRotation.Z)

	case mvFly:

		//the wind up key on blue gopher
		for _, anim := range mg.soloanims {
			anim.Update(0.001) //this interacts with the anim Speed()
		}

		//approach() applies smooth motions
		mg.vecRotation.X = Approach(mg.vecRotationGoal.X, mg.vecRotation.X, dtime/5)
		currentNode.RotateX(mg.vecRotation.X)
		mg.vecRotation.Y = Approach(mg.vecRotationGoal.Y, mg.vecRotation.Y, dtime/5)
		currentNode.RotateY(mg.vecRotation.Y)
		mg.vecRotation.Z = Approach(mg.vecRotationGoal.Z, mg.vecRotation.Z, dtime/5)
		currentNode.RotateZ(mg.vecRotation.Z)

		mg.vecMovement.SetX(Approach(mg.vecMovementGoal.X, mg.vecMovement.X, dtime))
		mg.vecMovement.SetY(Approach(mg.vecMovementGoal.Y, mg.vecMovement.Y, dtime))
		mg.vecMovement.SetZ(Approach(mg.vecMovementGoal.Z, mg.vecMovement.Z, dtime))

		//here is the gold nugget I got regarding flying / running around a room algorithm
		//see https://www.youtube.com/watch?v=FT7MShdqK6w&list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My&index=15

		//we need to calculate the two axes at 90 deg from the forward direction so we can apply trhust
		currentNode.WorldDirection(&vecViewForward)

		//see footnote1
		currentNode.WorldRotation(&vecViewUp)
		c := math32.Cos(vecViewUp.X) * math32.Cos(vecViewUp.Z)

		//thrusting/strafing calcs, to get the object's forward, up, right axes
		//broken in that positive/negative switch sometimes, I don't yet know why.
		// <<Jubilation after a lot of work and testing and failure>> =  Holy shit, it works! Mostly.
		vecViewUp.Set(
			math32.Cos(vecViewUp.Y)*c,
			math32.Sin(vecViewUp.Z),
			math32.Sin(vecViewUp.Y)*c)

		//2d games just need forward and right, Up (Y) can be gravity, just make sure char can't fall through floor
		vecViewForward.Normalize()
		vecViewUp.Normalize()
		vecViewUp.Cross(&vecViewForward)
		vecViewUp.Normalize()

		vecViewTmp.Copy(&vecViewUp) //Cross() modifies the vector so use a copy
		vecViewRight = *vecViewTmp.Cross(&vecViewForward)
		vecViewRight.Normalize()

		//apply the buffered (approach'd) movement to the vectors
		vecViewForward.MultiplyScalar(mg.vecMovement.Z)
		vecViewRight.MultiplyScalar(mg.vecMovement.X)
		vecViewUp.MultiplyScalar(mg.vecMovement.Y)

		//build velocity vector from everything above
		mg.vecVelocity = *vecViewForward.Add(&vecViewRight)
		mg.vecVelocity.Add(&vecViewUp)

		//finally apply the manipulated velocity to the position, et voila: motion
		usePos = currentNode.Position()
		currentNode.SetPositionVec(usePos.Add(&mg.vecVelocity))

		//gravity (notice it is placed on movement not velocity, it will be applied next frame):
		//symbolically mg.vecMovement = mg.vecMovement + mg.vecGravity * dtime;
		//g3n'd mg.vecMovement.Add(mg.vecGravity.MultiplyScalar(dtime))
		//since I didn't have a run and jump style demo I did not implement a gravity vector
		//above is how you would do it with a vecGravity like (0, -9.8, 0) where the -9.8
		//is earth's gravity attractive acceleration which will generally be in the Y axis but may be your Z
		//see https://www.youtube.com/watch?v=c4b9lCfSDQM&list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My&index=12
	}
}

// Game onKeyDown handler
func (gm *GameApp) onKeyDown(evname string, ev interface{}) {

	kev := ev.(*window.KeyEvent)
	switch kev.Key {

	case window.KeyF:
		gm.ToggleFullScreen()
		return

	case window.KeyQ:
		gm.Quit()
	}

	//send keystrokes to moveGopher
	demo.onKeyDown(gm, kev)

}

//moveGopher key  handler
func (mg *moveGopher) onKeyDown(gm *GameApp, kev *window.KeyEvent) {

	//kev := ev.(*window.KeyEvent)

	switch mvType {

	case mvTranslate:
		mg.Translate(kev)

	case mvFly:
		mg.Fly(kev)
	}

	switch kev.Key {

	case window.KeyB: //stop all rotations
		mg.vecRotation.Zero()
		mg.vecRotationPaused.Zero()
		mg.vecRotationGoal.Zero()

	case window.KeyD: //positive linear Approach sphere1
		mg.vecAppVelocityGoal.SetZ(0.2)

	case window.KeyE: //negative linear Approach sphere1
		mg.vecAppVelocityGoal.SetZ(-0.2)

	case window.KeyL: //LookAt's, direct and SLERP
		mg.getSlerpVector()

		//Control was pressed, use the Direct LookAt()
		if kev.Mods&window.ModControl > 0 {
			mg.soloGopher.LookAt(&vecLookAt, vecUpHat)
			return
		}

		//Control Key was not pressed, use the SLERP LookAt()
		mg.getSlerpQuats()
		go mg.quatSlerp(30, &mg.fromQuat, &mg.toQuat)
		mg.reset3DNormals()

	case window.KeyM: //toggle between Movement types: Translate vs Flying
		mg.doReset(gm)
		mvCnt++
		mvType = mvCnt % 2

	case window.KeyN: //flip Node between green gopher and camera

		switch nodeIsGopher {
		case true:
			gm.Camera.Add(gm.Ship)
			currentNode = gm.Camera.GetNode()
			nodeIsGopher = false
			//yet another disconnect between flying camera vs. green gopher
			//They are so different that it would be best probably to set those up as
			//separate movement routines.
			if mvType == mvFly {
				mg.vecMovementGoal.SetZ(mg.vecMovementGoal.Z * -1)
				gm.Log.Info("yesssss")
			}
		default:
			gm.Camera.Remove(gm.Ship)
			currentNode = mg.gopher
			nodeIsGopher = true
			if mvType == mvFly {
				mg.vecMovementGoal.SetZ(mg.vecMovementGoal.Z * -1)
			}
		}

	case window.Key0, window.KeyKP0, window.KeyO: //reset
		mg.doReset(gm)

	case window.KeyS: //stop all motion
		mg.stop()

	case window.KeyT: //toggle on/off
		mg.togglePause()

	}
}

//this sets the rotations that will be used in simple translation routine, called from onKey
func (mg *moveGopher) ChangeRotation(kev *window.KeyEvent) {

	incRot = incrementRotTranslate

	//Control Key decrements
	if kev.Mods&window.ModControl > 0 {
		incRot *= -1
	}

	switch kev.Key {

	case window.KeyX:
		mg.vecRotation.X += incRot

	case window.KeyY:
		mg.vecRotation.Y += incRot

	case window.KeyZ:
		mg.vecRotation.Z += incRot
	}
}

//this sets the motion vectors that will be used in simple translation, called from onKey
func (mg *moveGopher) Translate(kev *window.KeyEvent) {

	if kev.Mods&window.ModShift > 0 {
		mg.ChangeRotation(kev)
		return
	}

	incLinear = incrementLinear
	locAcceleration := incAcceleration

	//Control Key decrements velocity and acceleration
	if kev.Mods == window.ModControl {
		incLinear *= -1
		locAcceleration = 1 / locAcceleration
	}

	switch kev.Key {
	case window.KeyW:
		mg.vecVelocity.MultiplyScalar(locAcceleration)

		//these are local xyz/s, they move along the world axes. Rotations do not affect the linear
		//movement so good for modeling/moving satellites, thrown bottles, etc.
	case window.KeyX:
		mg.vecVelocity.X += incLinear

	case window.KeyY:
		mg.vecVelocity.Y += incLinear

	case window.KeyZ:
		mg.vecVelocity.Z += incLinear

	}
}

//this sets the motion vectors that will be used in flying using approach methods, called from onKey
func (mg *moveGopher) Fly(kev *window.KeyEvent) {

	incLinear = -incrementLinear
	locAcceleration := incAcceleration
	switch nodeIsGopher {
	case true:
		incRot = incrementRotFly
	default:
		incLinear *= -1
		incRot = incrementRotFly / 3
	}

	//Control Key decrements velocity, acceleration, and rotation
	if kev.Mods == window.ModControl {
		incLinear *= -1
		locAcceleration = 1 / locAcceleration
		incRot *= -1
	}

	switch kev.Key {

	//This applies a LARGE change so that the approach() can be easily seen
	case window.KeyA:
		if incRot < 0 {
			mg.vecRotationGoal.X += math32.Pi / 8
		} else {
			mg.vecRotationGoal.X += math32.Pi / -8
		}

	case window.KeyW:
		mg.vecVelocity.MultiplyScalar(locAcceleration)

	case window.KeyP:
		mg.vecRotationGoal.X += incRot

	case window.KeyY:
		mg.vecRotationGoal.Y += incRot

	case window.KeyR:
		mg.vecRotationGoal.Z += incRot

	case window.KeyZ:
		mg.vecMovementGoal.Z += -incLinear

	case window.KeyH: //horizontal thrust
		switch nodeIsGopher {
		case true:
			mg.vecMovementGoal.X += incLinear
		default:
			mg.vecMovementGoal.Y -= incLinear
		}

	case window.KeyV: //vertical thrust
		switch nodeIsGopher {
		case true:
			mg.vecMovementGoal.Y += incLinear
		default:
			mg.vecMovementGoal.X -= incLinear
		}
	}
}

//given from/to quaternion do a sherical linear interpolation, this is called as a go func,
//cnt is how many interpolations you want per given a ticker of 1/60 of a second,
//do NOT use this on an object that is being rendered in the render loop, you will not be happy
func (mg *moveGopher) quatSlerp(cnt float32, from, to *math32.Quaternion) {

	ticker := time.NewTicker(time.Millisecond * 34) //about 60 times a second
	//cnt := float32(30.0)

	for range ticker.C {
		//the Slerp() func changes the slerped quat. if you leave this alone the
		//slerping accelerates because the distance to rotate gets smaller and smaller.
		//this is my method to get a linear slerp, by using the inverse it adjusts for
		//the changing slerp length
		mg.soloGopher.SetRotationQuat(from.Slerp(to, 1/cnt))
		cnt--
		if cnt <= 0 {
			ticker.Stop()
			break
		}
	}
	//g.Log.Info("yeah you need the break statement")
}

//a not beautiful, quick/dirty info message
func (t *moveGopher) getCurrentInfo() string {
	var vecT1, vecT2, vecI, vec1, vec2 math32.Vector3

	vecT1 = t.sphere1.Position()
	vecT2 = t.sphere2.Position()
	vecI = currentNode.Position()
	vec1 = *vecT1.Sub(&vecI)
	vec2 = *vecT2.Sub(&vecI)

	//-----distance compare
	result := ""
	//fast comparison
	if vec1.LengthSq() >= vec2.LengthSq() {
		result = "big sphere closer\n"
	} else {
		result = "small sphere closer\n"
	}

	//-----BackStab
	//see theory https://www.youtube.com/watch?v=Q9FZllr6-wY&list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My&index=10
	//and code https://www.youtube.com/watch?v=HXpSQ7yyu3o&list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My&index=11
	vec1.Normalize()
	vec2.Normalize()

	currentNode.WorldDirection(&vecViewForward)
	if !nodeIsGopher {
		//continuing disconnects when working with camera vs. gopher object
		vecViewForward.MultiplyScalar(-1)
	}
	//vecViewForward.Normalize() //don't seem to need this

	if vecViewForward.Dot(&vec1) < -0.8 {
		result = result + "small sphere backstab!\n"
	}

	if vecViewForward.Dot(&vec2) < -0.8 {
		result = result + "big sphere backstab!\n"
	}

	//let's see what magnitude velocity is...
	result = result + fmt.Sprintf("vel: %v\n", t.vecVelocity.Length())

	return result
}

//This does an easein/easeout for motion and rotation, use the deltatime and
//divide it to get longer ramp, multiply to get faster ramp, this is not my
//creation, see link in code.
//https://www.youtube.com/watch?v=qJq7I2DLGzI&list=PLW3Zl3wyJwWOpdhYedlD-yCB7WQoHf-My&index=13
func Approach(goal, current, dtime float32) float32 {
	flDifference = goal - current
	if flDifference > dtime {
		return current + dtime
	}
	if flDifference < -dtime {
		return current - dtime
	}
	return goal
}

//Pause movement of the objects
func (mg *moveGopher) togglePause() {
	if mg.vecVelocity.Equals(&zeroVector) {
		mg.vecVelocity.Copy(&mg.vecVelocityPaused)
	} else {
		mg.vecVelocityPaused.Copy(&mg.vecVelocity)
		mg.vecVelocity.Zero()
	}

	switch mvType {
	case mvTranslate:

		if mg.vecRotation.Equals(&zeroVector) {
			mg.vecRotation.Copy(&mg.vecRotationPaused)
		} else {
			mg.vecRotationPaused.Copy(&mg.vecRotation)
			mg.vecRotation.Zero()
		}

	case mvFly:
		if mg.vecMovementGoal.Equals(&zeroVector) {
			mg.vecMovementGoal.Copy(&mg.vecMovementPaused)
		} else {
			mg.vecMovementPaused.Copy(&mg.vecMovementGoal)
			mg.vecMovementGoal.Zero()
		}

		if mg.vecRotationGoal.Equals(&zeroVector) {
			mg.vecRotationGoal.Copy(&mg.vecRotationPaused)
		} else {
			mg.vecRotationPaused.Copy(&mg.vecRotationGoal)
			mg.vecRotationGoal.Zero()
		}
	}

}

//Stop all movements, quat slerps in go routines will finish however,
//once stopped you can continue on by pressing the movement keys again
func (mg *moveGopher) stop() {
	mg.vecRotation.Zero()
	mg.vecRotationPaused.Zero()
	mg.vecRotationGoal.Zero()

	mg.vecVelocity.Zero()
	mg.vecVelocityPaused.Zero()

	mg.vecAppVelocity.Zero()
	mg.vecAppVelocityGoal.Zero()

	mg.vecMovement.Zero()
	mg.vecMovementGoal.Zero()
	mg.vecMovementPaused.Zero()
}

//Re-set all objects to start values
func (mg *moveGopher) doReset(gm *GameApp) {

	mg.stop()
	mg.sphere1.SetPosition(-10, 4, 10)
	currentNode = mg.gopher
	gm.Camera.Remove(gm.Ship)
	nodeIsGopher = true

	mg.reset3DNormals()

	mg.gopher.SetRotationVec(&zeroVector)
	mg.soloGopher.SetRotationVec(&zeroVector)
	mg.gopher.SetPosition(0, 0, 0)

	gm.Camera.SetRotationVec(&zeroVector)
	gm.Camera.SetPositionVec(&cameraVector)

	//Whoa! The flying thrusts Y/X get reversed, and smudged, if the object has used a LookAt!!!! Ouch.
	//I adjust by applying x to y, and vice versa, in the keystrokes. This needs to be worked on and understood.
	gm.Camera.LookAt(&zeroVector, vecUpHat)

}

//Does the work of changing the LookAt targets and calculating the
//"fixed" vector
func (mg *moveGopher) getSlerpVector() {
	mg.reset3DNormals()

	ToggleLookAtTarget++
	switch ToggleLookAtTarget % 3 {
	case 0:
		mg.sphere1.WorldPosition(&vecLookAtTarget)
	case 1:
		mg.sphere2.WorldPosition(&vecLookAtTarget)
	case 2:
		currentNode.WorldPosition(&vecLookAtTarget)
	}

	mg.soloGopher.WorldPosition(&vecLookAtLooker)

	//These calcs must be done for object lookAt's, there is a disconnect
	//with object LookAt's which use, I believe, a camera LookAt, which is wrong
	//for an object. I stumbled across this vector fix. Literally. I made a mistake
	//and subtracted twice when I meant to comment one out. Wow-ouch.

	//the subtraction order matters
	vecLookAt = *vecLookAtTarget.Sub(&vecLookAtLooker)
	vecLookAt.SubVectors(&vecLookAtLooker, &vecLookAtTarget)
}

//Does the work of getting the from and to quaternions needed for a smooth slerp
func (mg *moveGopher) getSlerpQuats() {

	rotMatrix.LookAt(&vecLookAtLooker, &vecLookAt, vecUpHat)

	mg.toQuat.SetFromRotationMatrix(&rotMatrix)
	mg.toQuat.Normalize() //whoops! Yes, you need to do this

	mg.fromQuat = mg.soloGopher.Quaternion()
	mg.fromQuat.Normalize() //whoops! Yes, you need to do this
}

/*
footnote1
-----------
look down z
y = sin Z
x = cos z

lookk down y
z = sin Y
x = cos Y

look down x
y = sin x <==?? This just doesn't work anywhere??
z = cos x

//convert rotation angles to vector
Vx = cos Y cos Z cos X  // (no, why? sinX)
Vy = sin Z              // (no, why? sinX)
Vz = sin Y cos X cos Z  // (no, why? sinX)

*/
