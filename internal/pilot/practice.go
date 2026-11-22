package pilot

import (
	"math"

	"zoro/internal/wire"
)

// tag names the stance for the debug label.
func (st stance) tag(ffa bool) string {
	switch {
	case st.duel:
		return "duel"
	case st.squad:
		return "squad"
	case ffa:
		return "ffa"
	}
	return "open"
}

// dropRun returns a supply drop worth breaking off for against the house squad: in reach, landing
// about when we would get there, and staying inside the zone. A drop is a full repair and cubes, a
// second life with the squad on us. nil when we are healthy with cubes, or nothing qualifies.
func (p *Pilot) dropRun(st stance, hpFrac float64) *wire.Drop {
	if p.T.DropRun <= 0 || !st.squad || (hpFrac > p.T.DropRunHP && p.o.Self.Cubes >= 10) {
		return nil
	}
	var best *wire.Drop
	bd := math.Inf(1)
	for i := range p.o.Drops {
		dr := &p.o.Drops[i]
		d := flatDist(p.pos, dr.Pos)
		if p.banned(dr.ID) || d > p.T.DropRun || dr.LandsIn > d/p.walk*1.25+p.T.DropRunLead {
			continue
		}
		if !p.safeLater(dr.Pos[0], dr.Pos[1], d/p.walk+6, 4) {
			continue
		}
		if d < bd {
			best, bd = dr, d
		}
	}
	return best
}

// seekTarget is where to go looking for the next fight against the squad when there is nothing else to
// do: the last place of the nearest mech seen lately that we would take on, else somewhere unexplored.
// Practice score is about kills, so time spent waiting for them to come is wasted.
func (p *Pilot) seekTarget(st stance, hpFrac float64) (x, y float64, id int, ok bool) {
	if p.T.Seek <= 0 || !st.squad || hpFrac < p.T.SeekHP {
		return 0, 0, -1, false
	}
	var best *Contact
	bd := math.Inf(1)
	for _, c := range p.contacts {
		if c.Drone || p.now-c.SeenT > p.T.Seek {
			continue
		}
		if u, banned := p.seekBans[c.ID]; banned && p.now < u {
			continue
		}
		d := flatDist(p.pos, c.Pos)
		if d > p.T.SeekRange || !p.safeLater(c.Pos[0], c.Pos[1], 8, 6) {
			continue
		}
		if d < 22 && !c.Listed {
			p.seekBans[c.ID] = p.now + 20 // we are where it was and it is not here
			continue
		}
		if adv, _, _ := p.odds(c, d); adv < st.engage {
			continue
		}
		if d < bd || (d == bd && best != nil && c.ID < best.ID) {
			best, bd = c, d
		}
	}
	if best != nil {
		return best.Pos[0], best.Pos[1], best.ID, true
	}
	x, y, ok = p.exploreGoal()
	return x, y, -1, ok
}

// jink zig-zags the heading while we walk somewhere under fire from range: shooters lead with our
// current velocity, so a straight walk away is the easiest target there is.
func (p *Pilot) jink(tn *turn) {
	lg := &tn.lg
	if p.T.Jink <= 0 || tn.direct || !tn.travelling || lg.still || lg.stayOnRoof || !tn.hurting || (lg.dirX == 0 && lg.dirY == 0) {
		return
	}
	if length(tn.goal[0]-p.pos[0], tn.goal[1]-p.pos[1]) <= 12 {
		return
	}
	far := false
	for i := range tn.ths {
		t := &tn.ths[i]
		if t.engaged && !t.ghost && t.d > 55 && p.now-t.c.HitMeT < 3 {
			far = true
		}
	}
	if !far {
		return
	}
	if p.now >= p.jinkT {
		if p.rng.Float64() < 0.85 {
			p.jinkDir = -p.jinkDir
		}
		p.jinkT = p.now + 0.3 + p.rng.Float64()*0.4
	}
	k := p.T.Jink * p.jinkDir
	lg.dirX, lg.dirY = unit(lg.dirX-lg.dirY*k, lg.dirY+lg.dirX*k)
}

// patient reports whether we are still building up against the squad (cubes and a laser first).
func (p *Pilot) patient(tn *turn) bool {
	return tn.st.squad && p.T.PatientSecs > 0 && p.now < p.T.PatientSecs &&
		(p.o.Self.Cubes < p.T.PatientCubes || p.holding("laser") < 0)
}

// threatField charges each candidate position for the extra damage we would take there over the
// next FieldWeight seconds, from the two sources that movement can actually avoid:
//   - a lightning carrier's reach (it does several times a gun's damage, but only up close), unless
//     we can answer it with lightning of our own;
//   - with two or more on us, an attacker left outside our front arc while we face the target, whose
//     hits then do rear damage (half as much again).
func (p *Pilot) threatField(tn *turn) {
	w := p.T.FieldWeight
	if w <= 0 {
		return
	}
	type arc struct{ x, y, reach, dps float64 }
	type flank struct{ x, y, extra float64 }
	var arcs []arc
	var flanks []flank
	lw := p.cfg.Weapon("lightning")
	canAnswer := p.holding("lightning") >= 0 && p.energy > 20
	for i := range tn.ths {
		t := &tn.ths[i]
		c := t.c
		if !c.Listed && p.now-c.SeenT > 2 {
			continue
		}
		if c.Arc && !canAnswer && t.d < 140 {
			reach := lw.Get("range", 45) + p.T.ArcMargin
			if c.Cubes >= 6 {
				reach += 5 // it may have bought the range upgrade
			}
			dps := lw.Get("damage", 60) / math.Max(lw.Get("cooldown", 0.25), 0.05) * p.damageMul(c.Cubes)
			arcs = append(arcs, arc{c.Pos[0], c.Pos[1], reach, dps})
			if t.d < reach+6 && c.Visible {
				tn.lg.mayRun = true
				tn.lg.look = max(tn.lg.look, 10)
			}
		}
		if tn.tgt != nil && tn.engaged >= 2 && t.engaged && c.ID != tn.tgt.c.ID && c.Visible && t.d < 220 {
			flanks = append(flanks, flank{c.Pos[0], c.Pos[1], 0.5 * p.theirDPS(c, t.d)})
		}
	}
	if len(arcs) == 0 && len(flanks) == 0 {
		return
	}
	half := p.mechStat("front_arc_deg", 180) * math.Pi / 360
	var tx, ty float64
	if tn.tgt != nil {
		tx, ty = tn.tgt.c.Pos[0], tn.tgt.c.Pos[1]
	}
	prev := tn.lg.extra
	tn.lg.extra = func(x, y float64) float64 {
		cost := 0.0
		if prev != nil {
			cost = prev(x, y)
		}
		for _, a := range arcs {
			// full price inside the reach, fading out over ten metres beyond it
			if f := clamp((a.reach+10-length(x-a.x, y-a.y))/10, 0, 1); f > 0 {
				cost += a.dps * w * f
			}
		}
		if len(flanks) > 0 {
			face := math.Atan2(ty-y, tx-x)
			for _, f := range flanks {
				if math.Abs(wrap(math.Atan2(f.y-y, f.x-x), face)) > half {
					cost += f.extra * w
				}
			}
		}
		return cost
	}
}

// killValue weights a target by how soon we could finish it (damage that does not kill within the
// regeneration window heals back) and by whether it has its back to us.
func (p *Pilot) killValue(t *threat) float64 {
	c := t.c
	rate := p.ourDPS(t.d)
	turned := p.T.Backstab > 0 && !c.Drone && math.Abs(wrap(c.Aim[0], bearing(c.Pos, p.pos))) > p.mechStat("front_arc_deg", 180)*math.Pi/360
	if turned {
		rate *= 1.5
	}
	v := 1.0
	switch {
	case p.T.KillTime > 0:
		hp := c.HP
		if hp <= 0 {
			hp = math.Max(c.MaxHP, 1)
		}
		v = 1 + p.T.KillTime/(hp/math.Max(rate, 1)+1)
	case c.MaxHP > 0:
		v = 2 - c.HP/c.MaxHP
	}
	if turned {
		v *= p.T.Backstab
	}
	return v
}

// laserLive reports whether c's laser could hurt us now. A laser announces its charge, so one that is
// not charging and fired less than a cooldown plus a charge ago cannot fire yet.
func (p *Pilot) laserLive(c *Contact) bool {
	if c.Charging {
		return true
	}
	w := p.cfg.Weapon("laser")
	return p.now-c.LaserT >= w.Get("cooldown", 1.65)+w.Get("charge", 0.75)
}

// endgame reports whether the late phase has come, when the whole squad hunts us at once.
func (p *Pilot) endgame() bool {
	z := &p.o.Zone
	return (p.T.EndgameT > 0 && p.now >= p.T.EndgameT) ||
		(p.T.EndgameR > 0 && z.NextRadius > 0 && z.NextRadius <= p.T.EndgameR && p.now >= 150)
}

// topUp reloads the gun below TopUp of a magazine whenever no recent threat can see us.
func (p *Pilot) topUp(cmd *wire.Command, tn *turn) {
	g := p.holding("gun")
	if p.T.TopUp <= 0 || g < 0 || cmd.Reload || cmd.Fire[g] {
		return
	}
	sl := &p.slots[g]
	if sl.reload > 0 || sl.ammo >= sl.stats.Get("mag", 12)*p.T.TopUp {
		return
	}
	for i := range tn.ths {
		if tn.ths[i].c.Visible && tn.ths[i].c.Listed {
			return
		}
	}
	if p.hiddenNow(tn.ths) {
		cmd.Reload = true
	}
}

// finishing reports whether the target is close enough to death to be worth closing in on.
func (p *Pilot) finishing(tn *turn) bool {
	t := tn.tgt
	if p.T.FinishHP <= 0 || t == nil || t.c.MaxHP <= 0 || !t.c.Visible || tn.engaged > 2 {
		return false
	}
	return t.c.HP < p.T.FinishHP*t.c.MaxHP && tn.hpFrac > 0.42
}

// armed reports whether one of our weapons could hurt the target within the next moment: a loaded
// gun in range, a laser near full charge, or lightning in reach with the energy for a few arcs.
func (p *Pilot) armedNow(d float64) bool {
	if g := p.holding("gun"); g >= 0 {
		sl := &p.slots[g]
		if sl.ammo > 0 && sl.reload <= 0.25 && d < sl.stats.Get("range", 230)-10 {
			return true
		}
	}
	if l := p.holding("laser"); l >= 0 {
		sl := &p.slots[l]
		if sl.charge >= sl.stats.Get("charge", 0.75)-0.3 && d < sl.stats.Get("range", 280)-5 {
			return true
		}
	}
	if a := p.holding("lightning"); a >= 0 {
		sl := &p.slots[a]
		if d < sl.stats.Get("range", 45)+5 && p.energy >= 3*sl.stats.Get("energy", 3) {
			return true
		}
	}
	return false
}

// coverCycle makes the fight movement use cover in rhythm with our weapons: while nothing of ours
// can fire, a spot the target can see costs CoverCycle; once we can fire, a spot it cannot see does.
// We take its shots only when we are trading.
func (p *Pilot) coverCycle(tn *turn) {
	t := tn.tgt
	w := p.T.CoverCycle
	if w <= 0 || t == nil || !t.c.Visible || p.finishing(tn) {
		return
	}
	ready := p.armedNow(t.d)
	e := eye{t.c.Pos[0], t.c.Pos[1], t.c.Pos[2] + 8*p.bodyScale(t.c.Cubes)}
	feet := p.pos[2]
	prev := tn.lg.extra
	tn.lg.extra = func(x, y float64) float64 {
		cost := 0.0
		if prev != nil {
			cost = prev(x, y)
		}
		if p.seenFrom(e, x, y, feet, 0.95) != ready {
			cost += w
		}
		return cost
	}
}
