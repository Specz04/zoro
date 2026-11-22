package pilot

// Tuning holds the behaviour constants. They are grouped by the part of the pilot that reads them.
type Tuning struct {
	Name string

	// fighting
	Engage        float64 // start a fight with someone not on us at this advantage or better
	Disengage     float64 // stance default for when a fight is lost
	DuelLeave     float64 // leave a one-on-one only below this advantage
	SwarmTTD      float64 // with several attackers, leave when this many seconds from death
	RetreatHP     float64 // ... or below this health share
	ReturnHP      float64 // come back out above this health share
	HuntHP        float64 // only go hunting above this health share
	HuntRange     float64 // hunt targets inside this range
	HostileMemory float64 // a mech that hurt us stays an attacker this long
	FarRange      float64 // beyond this an exchange without an edge is declined
	FarEdge       float64 // ... the advantage that counts as an edge
	BigEdge       float64 // an advantage so large the fight is as good as won
	CoverReach    float64 // path length we accept to reach cover
	CoverReachLas float64 // ... when the attacker holds a laser
	CoverMinRange float64 // inside this range we just fight
	EdgeMemory    float64 // seconds a fight/cover decision is held before it may flip
	QuietRange    float64 // a mech that has not touched us and is farther than this is left alone
	LightFireDPS  float64 // under lighter fire than this from range, keep farming
	LaserWatch    float64 // a laser looking at us counts as an attack out to this range
	KiteMin       float64 // range band held against two or more attackers
	KiteMax       float64
	ReturnFire    float64 // ducking on a roof: only shoot back inside this range
	GunMaxRange   float64
	LeadFloor     float64 // least share of a target's velocity trusted on long flights
	LeadEvidence  float64 // virtual shots needed before the learned lead is trusted
	StrafeMin     float64
	StrafeMax     float64

	// cover after the first hidden spot
	DeepMargin float64 // cover must survive the threat side-stepping this far
	DeepRange  float64 // hurt with the attacker inside this range: keep going deeper
	DeepGain   float64 // ... each leg must add this much distance

	// looting
	LootRadius  float64
	WantCubes   int
	HiddenBias  float64 // loot in a known mech's sight is worth this much less
	ArmSwapMin  float64 // replace a held weapon only for this much more value
	LaserValue  float64
	LightValue  float64
	LaserSprint float64 // sprint to a laser on the ground above this energy
	ZoneSlack   float64 // loot must stay in the zone this long after we arrive
	DropRadius  float64
	DropGreed   float64
	CrateGiveUp float64 // seconds at a crate without hurting it before we leave it alone
	DroneGiveUp float64
	QuietHold   float64 // seconds without watchers before we call it quiet

	// roofs
	UseRoofs      bool
	RoofMaxClimb  float64
	RoofMinHeight float64
	RoofMinZone   float64
	RoofCubes     int     // wait on a roof once we have a laser, this many cubes ...
	RoofAfter     float64 // ... or this much match time has passed
	RoofDuckHold  float64 // stay behind the roof edge this long after the last watcher
	RoofLeaveCalm float64 // leave a doomed roof only when unwatched until the zone is this close (s)
	ClimbClear    float64 // no climbing with a mech in plain sight inside this range
}

// Defaults returns the tuning the bot plays with.
func Defaults() Tuning {
	return Tuning{
		Name: "zoro",

		Engage:        1.3,
		Disengage:     0.75,
		DuelLeave:     0.5,
		SwarmTTD:      10,
		RetreatHP:     0.45,
		ReturnHP:      0.85,
		HuntHP:        0.6,
		HuntRange:     130,
		HostileMemory: 4,
		FarRange:      85,
		FarEdge:       1.25,
		BigEdge:       3.0,
		CoverReach:    55,
		CoverReachLas: 90,
		CoverMinRange: 40,
		EdgeMemory:    2.5,
		QuietRange:    40,
		LightFireDPS:  60,
		LaserWatch:    285,
		KiteMin:       130,
		KiteMax:       190,
		ReturnFire:    230,
		GunMaxRange:   224,
		LeadFloor:     0.65,
		LeadEvidence:  8,
		StrafeMin:     0.35,
		StrafeMax:     1.1,

		DeepMargin: 8,
		DeepRange:  80,
		DeepGain:   18,

		LootRadius:  220,
		WantCubes:   10,
		HiddenBias:  0.5,
		ArmSwapMin:  1.0,
		LaserValue:  6.0,
		LightValue:  2.6,
		LaserSprint: 30,
		ZoneSlack:   10,
		DropRadius:  300,
		DropGreed:   0.5,
		CrateGiveUp: 5,
		DroneGiveUp: 8,
		QuietHold:   3,

		UseRoofs:      true,
		RoofMaxClimb:  40,
		RoofMinHeight: 12,
		RoofMinZone:   60,
		RoofCubes:     6,
		RoofAfter:     60,
		RoofDuckHold:  6,
		RoofLeaveCalm: 25,
		ClimbClear:    150,
	}
}
