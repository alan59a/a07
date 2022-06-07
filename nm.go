// I couldn't get to make the optimize package in gonum to work ... so i got pissed ... so here is a new simpler BUT probably slower one
// I know there are multiple variations of the algorithm ... so i'm using  the wikipedia one ...
// I'm a Doctor ... not a mathematician ... so if i'm wrong please leave a note :D

// and one last thing ... this thing does NOT stop ... choose a bad function and it may go forever ... bad means without min/max

package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// It's a direct search (without derivatives) method to find local minimum/maximum
// and uses a simplex (shape with n+1 vertecis) ...
type NelderMead struct {
	ref float64 // reflection > 0 ... default = 1
	exp float64 // expansion > 1 ... default = 2
	con float64 // contraction > 0 but < 0.5 ... default = 0.5
	shr float64 // shrinkage ... default = 0.5
	fun func(input ...float64) float64
	dim int
	max bool

	// Choosing the correct min and max range of initial simplex vertecies may reduce computation time and increase final answer
	// Remember that choosing a small range may lead to a local min/max
	rmin []float64
	rmax []float64
	simp []simp

	// containing each step's centroid
	cen []float64

	// containing each step's standard deviasion
	stan float64

	// Standard Deviasion threshold for termination ... default = 0.0005
	thr float64

	// threshold for making sure the simplex is matured enough ... totally original ... alan made ...
	mat int

	// if the program has improved or not
	imp bool
}

// even i don't know what i'm doing ... lol,
// But you can try this:
// I'm making a simplex container so it can be sorted alongside the points resulting the values
// the simplex struct is a vertex in the our main simp
type simp struct {
	Vertex []float64
	Value  float64
}

func main() {
	fun := func(x ...float64) float64 {
		return math.Pow(x[0], 2) + math.Pow((x[1]+5), 2) + 10
	}
	nm := NewNM(fun, 2, false, []float64{-5, -5}, []float64{5, 5})
	a, b := nm.Converge()
	fmt.Println(a, b)
}

// If you do NOT want the default values so set them yourself ... lazy ...
func NewNM(function func(input ...float64) float64, dimensions int, max bool, rmin, rmax []float64) *NelderMead {

	// the simplex container ...
	sim := make([]simp, dimensions+1)

	// let's choose a few random points in our problem ...
	rand.Seed(time.Now().UnixNano())

	for a := 0; a <= dimensions; a++ { // amount of points needed for a simplex is dimensions + 1
		p := make([]float64, dimensions)
		for b := 0; b < dimensions; b++ {
			p[b] = rmin[b] + rand.Float64()*(rmax[b]-rmin[b])
		}

		// after setting a random point we should find out the outcome of the function
		// and i don't like to set a value to zero
		sim[a] = simp{
			Vertex: p,
			Value:  function(p...),
		}
		// and we'll sort them later
	}

	return &NelderMead{
		ref:  1,
		exp:  1.5,
		con:  0.5,
		shr:  0.5,
		fun:  function,
		dim:  dimensions,
		max:  max,
		rmin: rmin,
		rmax: rmax,
		simp: sim,
		stan: 1,
		thr:  0.0005,
		mat:  5,
		imp:  false,
	}
}

func (nm *NelderMead) Set(reflection, expantion, contraction, shrinkage, threshold float64, maturity int) {
	nm.ref = reflection
	nm.exp = expantion
	nm.con = contraction
	nm.shr = shrinkage
	nm.thr = threshold
	nm.mat = maturity
}

func (nm *NelderMead) Converge() (simp, int) {
	i := 0
	for {
		if nm.stan < nm.thr {
			nm.mat--
			if nm.mat == 0 {
				break
			}
		}
		i++
		if nm.imp {
			nm.imp = false
		}

		if nm.rerflect() {

			// The true reflect() output means that the reflection result was better that the best point
			// ... so we take our chances and try to expand:
			nm.expand()
		}

		// imp is switch so for each iteration and if there was no progression and contraction or shrinkage was needed ...
		if !nm.imp {
			if nm.contract() {
				nm.shrink()
			}
		}

		nm.sd()
		fmt.Printf("\r%4.2f", nm.simp[0].Value)
	}

	if nm.max {
		return simp{
			Vertex: nm.simp[len(nm.simp)-1].Vertex,
			Value:  nm.simp[len(nm.simp)-1].Value,
		}, i
	}

	return simp{
		Vertex: nm.simp[0].Vertex,
		Value:  nm.simp[0].Value,
	}, i
}

// REFLECTION PHASE
func (nm *NelderMead) rerflect() bool {
	// always make sure the simplex is sorted
	nm.sort()

	// calculate the centorid of the simplex ... of the intermediate and best verteces
	cen := nm.cens()

	ref := make([]float64, nm.dim)

	// try the reflection of the worst (either min or max) vertex
	if nm.max {

		for a := range ref {
			ref[a] = cen[a] + nm.ref*(cen[a]-nm.simp[0].Vertex[a])
		}

		res := nm.fun(ref...)

		if res > nm.simp[len(nm.simp)-1].Value {
			nm.simp[0].Value = res
			nm.simp[0].Vertex = ref

			nm.imp = true

			return false

		} else {

			for a := range nm.simp {

				if res > nm.simp[a].Value {
					nm.simp[0].Value = res
					nm.simp[0].Vertex = ref

					nm.imp = true

					return false
				}
			}

			return false
		}

	} else {

		for a := range ref {
			ref[a] = cen[a] + nm.ref*(cen[a]-nm.simp[len(nm.simp)-1].Vertex[a])
		}

		res := nm.fun(ref...)

		if res < nm.simp[0].Value {
			nm.simp[len(nm.simp)-1].Value = res
			nm.simp[len(nm.simp)-1].Vertex = ref

			nm.imp = true

			return true

		} else {

			for a := range nm.simp {

				if res < nm.simp[a].Value {
					nm.simp[len(nm.simp)-1].Value = res
					nm.simp[len(nm.simp)-1].Vertex = ref

					nm.imp = true

					return false
				}
			}
		}
	}
	return false
}

// EXPANSION PHASE
// This is used when the reflected point is the best point founded so far
func (nm *NelderMead) expand() {
	ref := make([]float64, nm.dim)

	if nm.max {

		for a := range ref {
			ref[a] = nm.cen[a] + nm.exp*(nm.cen[a]-nm.simp[0].Vertex[a])
		}

		res := nm.fun(ref...)

		if res > nm.simp[0].Value {
			nm.simp[0].Value = res
			nm.simp[0].Vertex = ref
		}

	} else {

		for a := range ref {
			ref[a] = nm.cen[a] + nm.exp*(nm.cen[a]-nm.simp[len(nm.simp)-1].Vertex[a])
		}

		res := nm.fun(ref...)

		if res < nm.simp[0].Value {
			nm.simp[0].Value = res
			nm.simp[0].Vertex = ref
		}
	}
}

// CONTRACTION PHASE
// This is used when the reflected point is no improvment at all
func (nm *NelderMead) contract() bool {
	c := make([]float64, nm.dim)

	if nm.max {

		for a := range c {
			c[a] = nm.cen[a] + nm.con*(nm.simp[0].Vertex[a]-nm.cen[a])
		}

		res := nm.fun(c...)

		if res > nm.simp[0].Value {
			nm.simp[0].Vertex = c
			nm.simp[0].Value = res

			return false
		}

		return true

	} else {

		for a := range c {
			c[a] = nm.cen[a] + nm.con*(nm.simp[len(nm.simp)-1].Vertex[a]-nm.cen[a])
		}

		res := nm.fun(c...)

		if res < nm.simp[len(nm.simp)-1].Value {
			nm.simp[len(nm.simp)-1].Vertex = c
			nm.simp[len(nm.simp)-1].Value = res

			return false
		}

		return true
	}
}

// SHRIKAGE PHASE
// This is the last straw ... replacing all bad vertecies with another one
func (nm *NelderMead) shrink() {
	if nm.max {

		for a := 0; a < nm.dim; a++ {

			for b := 0; b < len(nm.simp)-2; b++ {
				nm.simp[b].Vertex[a] = nm.simp[b].Vertex[a] + nm.con*(nm.simp[b].Vertex[a]-nm.simp[len(nm.simp)-1].Vertex[a])
			}
		}

	} else {

		for a := 0; a < nm.dim; a++ {

			for b := 1; b < len(nm.simp)-1; b++ {
				nm.simp[b].Vertex[a] = nm.simp[b].Vertex[a] + nm.con*(nm.simp[b].Vertex[a]-nm.simp[0].Vertex[a])
			}
		}
	}
}

func (nm *NelderMead) sort() {

	// This is actually a stupid way to sort a slice ... i know ... but again i'm lazy
	// in this sorting system a lower index means lower value result of the function
	for a := 0; a < nm.dim; a++ {

		for b := 0; b < nm.dim; b++ {

			if nm.simp[b].Value > nm.simp[b+1].Value {
				temp := nm.simp[b]
				nm.simp[b] = nm.simp[b+1]
				nm.simp[b+1] = temp
			}
		}
	}
}

// Standard Deviation
func (nm *NelderMead) sd() {
	sum := 0.0
	sd := 0.0

	for _, a := range nm.simp {
		sum += a.Value
	}

	for _, a := range nm.simp {
		sd += math.Pow((sum/float64(len(nm.simp)) - a.Value), 2)
	}

	nm.stan = sd
}

func (nm *NelderMead) cens() []float64 {
	switch len(nm.simp) {
	case 2:
		x1 := 0.0

		if nm.max {
			x1 += nm.simp[0].Vertex[0]
		} else {
			x1 += nm.simp[1].Vertex[0]
		}

		nm.cen = []float64{x1}

		return nm.cen

	case 3:
		x1 := 0.0
		x2 := 0.0

		if nm.max {

			for a := 1; a <= nm.dim; a++ {
				x1 += nm.simp[a].Vertex[0]
				x2 += nm.simp[a].Vertex[1]
			}

		} else {

			for a := 0; a < nm.dim; a++ {
				x1 += nm.simp[a].Vertex[0]
				x2 += nm.simp[a].Vertex[1]
			}

		}
		nm.cen = []float64{x1 / 2, x2 / 2}

		return nm.cen

	case 4:
		x1 := 0.0
		x2 := 0.0
		x3 := 0.0

		if nm.max {

			for a := 1; a <= nm.dim; a++ {
				x1 += nm.simp[a].Vertex[0]
				x2 += nm.simp[a].Vertex[1]
				x3 += nm.simp[a].Vertex[2]
			}

		} else {

			for a := 0; a < nm.dim; a++ {
				x1 += nm.simp[a].Vertex[0]
				x2 += nm.simp[a].Vertex[1]
				x3 += nm.simp[a].Vertex[2]
			}
		}

		nm.cen = []float64{x1 / 3, x2 / 3, x3 / 3}

		return nm.cen

	case 5:
		x1 := 0.0
		x2 := 0.0
		x3 := 0.0
		x4 := 0.0

		if nm.max {

			for a := 1; a <= nm.dim; a++ {
				x1 += nm.simp[a].Vertex[0]
				x2 += nm.simp[a].Vertex[1]
				x3 += nm.simp[a].Vertex[2]
				x4 += nm.simp[a].Vertex[3]
			}

		} else {

			for a := 0; a < nm.dim; a++ {
				x1 += nm.simp[a].Vertex[0]
				x2 += nm.simp[a].Vertex[1]
				x3 += nm.simp[a].Vertex[2]
				x4 += nm.simp[a].Vertex[3]
			}
		}

		nm.cen = []float64{x1 / 4, x2 / 4, x3 / 4, x4 / 4}

		return nm.cen

	default:
		log.Fatalln("probably not supported yet... ")
	}

	return nm.cen
}
