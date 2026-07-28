package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ---- depth constants (Z layers) ----
const (
	DepthGuiText     = 0
	DepthGui         = 1
	DepthShark       = 2
	DepthFishStart   = 3
	DepthFishEnd     = 20
	DepthSeaweed     = 21
	DepthCastle      = 22
	DepthWaterLine3  = 2
	DepthWaterGap3   = 3
	DepthWaterLine2  = 4
	DepthWaterGap2   = 5
	DepthWaterLine1  = 6
	DepthWaterGap1   = 7
	DepthWaterLine0  = 8
	DepthWaterGap0   = 9
)

// ===================== ENVIRONMENT =====================

func addEnvironment(eng *Engine) {
	segs := []string{
		"~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~",
		"^^^^ ^^^  ^^^   ^^^    ^^^^      ",
		"^^^^      ^^^^     ^^^    ^^     ",
		"^^      ^^^^      ^^^    ^^^^^^  ",
	}
	segLen := len(segs[0])
	rep := eng.Width()/segLen + 1
	for i, seg := range segs {
		tiled := strings.Repeat(seg, rep)
		frames, w, h := ParseFrames([]string{tiled})
			var zDepth int
			switch i {
			case 0:
				zDepth = DepthWaterLine0
			case 1:
				zDepth = DepthWaterLine1
			case 2:
				zDepth = DepthWaterLine2
			case 3:
				zDepth = DepthWaterLine3
			}

			eng.AddEntity(&Entity{
				Name:      fmt.Sprintf("water_seg_%d", i),
				Type:      "waterline",
				Frames:    frames,
				Width:     w,
				Height:    h,
				X:         0,
				Y:         float64(i + 5),
				Z:         zDepth,
			DefStyle:  namedStyle("cyan"),
			Physical:  true,
			CollDepth: 22,
		})
	}
}

// ===================== CASTLE =====================

func addCastle(eng *Engine) {
	shape := `
               T~~
               |
              /^\
             /   \
 _   _   _  /     \  _   _   _
[ ]_[ ]_[ ]/ _   _ \[ ]_[ ]_[ ]
|_=__-_ =_|_[ ]_[ ]_|_=-___-__|
 | _- =  | =_ = _    |= _=   |
 |= -[]  |- = _ =    |_-=_[] |
 | =_    |= - ___    | =_ =  |
 |=  []- |-  /| |\   |=_ =[] |
 |- =_   | =| | | |  |- = -  |
 |_______|__|_|_|_|__|_______|`

	mask := `
                 RR

               yyy
              y   y
             y     y
            y       y



               yyy
              yy yy
             y y y y
             yyyyyyy`

	frames, w, h := ParseFrames([]string{shape})
	colorFrames := ParseColorFrames([]string{mask})
	eng.AddEntity(&Entity{
		Name:        "castle",
		Frames:      frames,
		ColorFrames: colorFrames,
		Width:       w,
		Height:      h,
		X:           float64(eng.Width() - 32),
		Y:           float64(eng.Height() - 13),
		Z:           DepthCastle,
		DefStyle:    namedStyle("BLACK"),
	})
}

// ===================== SEAWEED =====================

func addAllSeaweed(eng *Engine) {
	count := eng.Width() / 15
	for i := 0; i < count; i++ {
		addSeaweed(nil, eng)
	}
}

func addSeaweed(_ *Entity, eng *Engine) {
	h := rand.Intn(4) + 3
	var frame0, frame1 strings.Builder
	for i := 1; i <= h; i++ {
		if i%2 == 1 {
			frame0.WriteString(" )\n")
			frame1.WriteString("(\n")
		} else {
			frame0.WriteString("(\n")
			frame1.WriteString(" )\n")
		}
	}
	frames, w, fh := ParseFrames([]string{frame0.String(), frame1.String()})
	x := float64(rand.Intn(eng.Width()-2) + 1)
	y := float64(eng.Height() - fh)
	speed := rand.Float64()*0.05 + 0.25

	eng.AddEntity(&Entity{
		Name:      fmt.Sprintf("seaweed_%d", rand.Int()),
		Frames:    frames,
		Width:     w,
		Height:    fh,
		X:         x,
		Y:         y,
		Z:         DepthSeaweed,
		AnimSpeed: speed,
		DefStyle:  namedStyle("green"),
		TransChar: ' ',
		DieTime:   time.Now().Unix() + int64(rand.Intn(4*60)) + 8*60,
		DeathCB:   addSeaweed,
	})
}

// ===================== BUBBLES =====================

func addBubble(fish *Entity, eng *Engine) {
	bx := fish.X
	if fish.DX > 0 {
		bx += float64(fish.Width)
	}
	by := fish.Y + float64(fish.Height/2)
	bz := fish.Z - 1

	frames, w, h := ParseFrames([]string{".", "o", "O", "O", "O"})
	eng.AddEntity(&Entity{
		Type:         "bubble",
		Frames:       frames,
		Width:        w,
		Height:       h,
		X:            bx,
		Y:            by,
		Z:            bz,
		DY:           -1,
		AnimSpeed:    0.1,
		DieOffscreen: true,
		Physical:     true,
		CollHandler:  bubbleCollision,
		DefStyle:     namedStyle("CYAN"),
		TransChar:    0,
	})
}

func bubbleCollision(bubble *Entity, eng *Engine) {
	for _, c := range bubble.Collisions {
		if c.Type == "waterline" {
			bubble.Dead = true
			return
		}
	}
}

// ===================== FISH =====================

// fishPair holds shape + color mask for one fish orientation.
type fishPair struct {
	shape string
	mask  string
}

func addAllFish(eng *Engine) {
	screenSize := (eng.Height() - 9) * eng.Width()
	if screenSize < 350 {
		screenSize = 350
	}
	count := screenSize / 350
	for i := 0; i < count; i++ {
		addFish(nil, eng)
	}
}

func addFish(_ *Entity, eng *Engine) {
	if rand.Intn(12) > 8 {
		addNewFish(eng)
	} else {
		addOldFish(eng)
	}
}

func addFishEntity(eng *Engine, pairs []fishPair) {
	fishNum := rand.Intn(len(pairs))
	p := pairs[fishNum]

	speed := rand.Float64()*2 + 0.25
	d := rand.Intn(DepthFishEnd-DepthFishStart) + DepthFishStart

	mask := strings.ReplaceAll(p.mask, "4", "W")
	mask = randColor(mask)

	if fishNum%2 == 1 {
		speed *= -1
	}

	frames, w, h := ParseFrames([]string{p.shape})
	colorFrames := ParseColorFrames([]string{mask})
	trans := DetectTransparent(frames)

	maxY := 9
	minY := eng.Height() - h
	if minY <= maxY {
		minY = maxY + 1
	}
	y := rand.Intn(minY-maxY) + maxY

	var x int
	if fishNum%2 == 1 {
		x = eng.Width() - 2
	} else {
		x = 1 - w
	}

	eng.AddEntity(&Entity{
		Type:         "fish",
		Frames:       frames,
		ColorFrames:  colorFrames,
		Width:        w,
		Height:       h,
		X:            float64(x),
		Y:            float64(y),
		Z:            d,
		DX:           speed,
		TransChar:    trans,
		DefStyle:     namedStyle("white"),
		Callback:     fishCallback,
		DieOffscreen: true,
		DeathCB:      addFish,
		Physical:     true,
		CollHandler:  fishCollision,
	})
}

func fishCallback(e *Entity, eng *Engine) {
	if rand.Intn(100) > 97 {
		addBubble(e, eng)
	}
	e.X += e.DX
	e.Y += e.DY
}

func fishCollision(fish *Entity, eng *Engine) {
	for _, c := range fish.Collisions {
		if c.Type == "teeth" && fish.Height <= 5 {
			addSplat(eng, c.X, c.Y, c.Z)
			fish.Dead = true
			return
		}
	}
}

func addSplat(eng *Engine, x, y float64, z int) {
	splatFrames := []string{
		"\n\n   .\n  ***\n   '\n",
		"\n\n \",*;`\n \"*,**\n *\"'~'\n",
		"  , ,\n \" \",\"'\n *\" *'\"\n  \" ; .\n",
		"* ' , ' `\n' ` * . '\n ' `' \",'\n* ' \" * .\n\" * ', '",
	}
	frames, w, h := ParseFrames(splatFrames)
	eng.AddEntity(&Entity{
		Type:      "splat",
		Frames:    frames,
		Width:     w,
		Height:    h,
		X:         x - 4,
		Y:         y - 2,
		Z:         z - 2,
		AnimSpeed: 0.25,
		DieFrame:  15,
		DefStyle:  namedStyle("RED"),
		TransChar: ' ',
	})
}

// ===================== OLD FISH SHAPES =====================

func addOldFish(eng *Engine) {
	pairs := []fishPair{
		// ---- type 0: right-facing large ----
		{replaceBackticks("       \\\n     ...\\.,\n\\  /'       \\\n >=     (  ' >\n/  \\      / /\n    §\"'\"'/''"), "       2\n     1112111\n6  11       1\n 66     7  4 5\n6  1      3 1\n    11111311"},
		// ---- type 1: left-facing large ----
		{replaceBackticks("      /\n  ,../...\n /       '\\  /\n< '  )     =<\n \\ \\      /  \\\n  §'\\'\"'\"'"), "      2\n  1112111\n 1       11  6\n5 4  7     66\n 1 3      1  6\n  11311111"},
		// ---- type 2: right-facing medium ----
		{"    \\\n\\ /--\\\n>=  (o>\n/ \\__/\n    /", "    2\n6 1111\n66  745\n6 1111\n    3"},
		// ---- type 3: left-facing medium ----
		{"  /\n /--\\ /\n<o)  =<\n \\__/ \\\n  \\", "  2\n 1111 6\n547  66\n 1111 6\n  3"},
		// ---- type 4: right-facing stripy ----
		{replaceBackticks("       \\:.\n\\;,   ,;\\\\\\,, \n  \\\\\\;;:::::::o\n  ///;;::::::::< \n /;§ §§/////§§"), "       222\n666   1122211\n  6661111111114\n  66611111111115\n 666 113333311"},
		// ---- type 5: left-facing stripy ----
		{replaceBackticks("      .:/\n   ,,///;,   ,;/\n o:::::::;;///\n>::::::;;\\\\\\\\\\\n  §§\\\\\\\\\\\\\\\\\\§§ §;\\ "), "      222\n   1122211   666\n 4111111111666\n51111111111666\n  113333311 666"},
		// ---- type 6: tiny right ----
		{"  __\n><_'>\n   '", "  11\n61145\n   3"},
		// ---- type 7: tiny left ----
		{" __\n<'_><\n `", " 11\n54116\n 3"},
		// ---- type 8: small right ----
		{replaceBackticks("   ..\\,\n>=§   ('\n  §§§/'§§"), "   1121\n661   745\n  111311"},
		// ---- type 9: small left ----
		{replaceBackticks("  ,/..\n<')   §=<\n §§\\§§§"), "  1211\n547   166\n 113111"},
		// ---- type 10: right mini ----
		{"   \\\n  / \\\n>=_('>\n  \\_/\n   /", "   2\n  1 1\n661745\n  111\n   3"},
		// ---- type 11: left mini ----
		{"  /\n / \\\n<')_=<\n \\_/\n  \\", "  2\n 1 1\n547166\n 111\n  3"},
		// ---- type 12: right tiny2 ----
		{"  ,\\\n>=('>\n  '/", "  12\n66745\n  13"},
		// ---- type 13: left tiny2 ----
		{" /,\n<')=<\n \\`", " 21\n54766\n 31"},
		// ---- type 14: right puffer ----
		{"  __\n\\/ o\\\n/\\__/", "  11\n61 41\n61111"},
		// ---- type 15: left puffer ----
		{" __\n/o \\/\n\\__/\\", " 11\n14 16\n11116"},
	}
	addFishEntity(eng, pairs)
}

// ===================== NEW FISH SHAPES =====================

func addNewFish(eng *Engine) {
	pairs := []fishPair{
		// ---- type 0: right ----
		{"   \\\n  / \\\n>=_('>\n  \\_/\n   /",
			"   1\n  1 1\n663745\n  111\n   3"},
		// ---- type 1: left ----
		{"  /\n / \\\n<')_=<\n \\_/\n  \\",
			"  2\n 111\n547366\n 111\n  3"},
		// ---- type 2: right fancy ----
		{replaceBackticks("     ,\n     }\\\n\\  .'  §\\\n}}<   ( 6>\n/  §,  .'\n     }/\n     '"),
			"     2\n     22\n6  11  11\n661   7 45\n6  11  11\n     33\n     3"},
		// ---- type 3: left fancy ----
		{replaceBackticks("    ,\n   /{\n /'  §.  /\n<6 )   >{{ \n §.  ,'  \\\n   \\{\n    §"),
			"    2\n   22\n 11  11  6\n54 7   166\n 11  11  6\n   33\n    3"},
		// ---- type 4: right big ----
		{replaceBackticks("            \\'§.\n             )  \\\n(§.??????_.-§' ' '§-.\n \\ §.??.§        (o) \\_\n  >  ><     (((       (\n / .§??§._      /_|  /'\n(.§???????§-. _  _.-§\n            /__/'"),
			"            1111\n             1  1\n111      11111 1 1111\n 1 11  11        141 11\n  1  11     777       5\n 1 11  111      333  11\n111       111 1  1111\n            11111"},
		// ---- type 5: left big ----
		{replaceBackticks("       .'§/\n      /  (\n  .-'§ § §'-._??????.')\n_/ (o)        '.??.' /\n)       )))     ><  <\n§\\  |_\\      _.'??'. \\\n  '-._  _ .-'???????'.)\n      §\\__\\"),
			"       1111\n      1  1\n  1111 1 11111      111\n11 141        11  11 1\n5       777     11  1\n11  333      111  11 1\n  1111  1 111       111\n      11111"},
		// ---- type 6: right puffer2 ----
		{replaceBackticks("       ,--,_\n__    _\\.---'-.\n\\ '.-\"     // o\\\n/_.'-._    \\\\  /\n       §\"--(/\"§"),
			"       22222\n66    121111211\n6 6111     77 41\n6661111    77  1\n       11113311"},
		// ---- type 7: left puffer2 ----
		{replaceBackticks("    _,--,\n .-'---./_    __\n/o \\\\     \"-.' /\n\\  //    _.-'._\\\n §\"\\)--\"§"),
			"    22222\n 112111121    66\n14 77     1116 6\n1  77    1111666\n 11331111"},
	}
	addFishEntity(eng, pairs)
}

// ===================== SHARK =====================

func addShark(_ *Entity, eng *Engine) {
	sharkShape := []string{
		replaceBackticks("                              __\n                             ( §\\\n  ,??????????????????????????)   §\\\n;' §.????????????????????????(     §\\__\n ;   §.?????????????__..---''          §~~~~-._\n  §.   §.____...--''                       (b  §--._\n    >                     _.-'      .((      ._     )\n  .§.-§--...__         .-'     -.___.....-(|/|/|/|/'\n ;.'?????????§. ...----§.___.',,,_______......---'\n '???????????'-'"),
		replaceBackticks("                     __\n                    /' )\n                  /'   (??????????????????????????,\n              __/'     )????????????????????????.' §;\n      _.-~~~~'          §---..__?????????????.'   ;\n _.--'  b)                       §--...____.'   .'\n(     _.      )).      §-._                     <\n §\\|\\|\\|\\|)-.....___.-     §-.         __...--'-.'.\n   §---......_______,,,§.___.'----... .'?????????§.;\n                                     §-§???????????§"),
	}
	sharkMask := []string{
		"\n\n\n\n\n                                           cR\n \n                                          cWWWWWWWW",
		"\n\n\n\n\n        Rc\n\n  WWWWWWWWc",
	}

	dir := rand.Intn(2)
	x := -53.0
	diffY := eng.Height() - 19
	if diffY <= 0 {
		diffY = 1
	}
	y := float64(rand.Intn(diffY) + 9)
	teethX := -9.0
	teethY := y + 7
	speed := 2.0
	if dir == 1 {
		speed = -2
		x = float64(eng.Width() - 2)
		teethX = x + 9
	}

	// Invisible "teeth" entity for collision with fish
	tFrames, tw, th := ParseFrames([]string{"*"})
	eng.AddEntity(&Entity{
		Type:      "teeth",
		Frames:    tFrames,
		Width:     tw,
		Height:    th,
		X:         teethX,
		Y:         teethY,
		Z:         DepthShark + 1,
		DX:        speed,
		CollDepth: DepthFishEnd - DepthFishStart,
		Physical:  true,
	})

	frames, w, h := ParseFrames([]string{sharkShape[dir]})
	colorFrames := ParseColorFrames([]string{sharkMask[dir]})
	trans := DetectTransparent(frames)

	eng.AddEntity(&Entity{
		Type:         "shark",
		Frames:       frames,
		ColorFrames:  colorFrames,
		Width:        w,
		Height:       h,
		X:            x,
		Y:            y,
		Z:            DepthShark,
		DX:           speed,
		TransChar:    trans,
		DieOffscreen: true,
		DeathCB:      sharkDeath,
		DefStyle:     namedStyle("CYAN"),
	})
}

func sharkDeath(_ *Entity, eng *Engine) {
	for _, e := range eng.GetEntitiesOfType("teeth") {
		e.Dead = true
	}
	randomObject(eng)
}

// ===================== SHIP =====================

func addShip(_ *Entity, eng *Engine) {
	shipShape := []string{
		"     |    |    |\n    )_)  )_)  )_)\n   )___))___))___)\\\n  )____)____)_____)\\\\  \n_____|____|____|____\\\\\\\n\\                   /",
		"         |    |    |\n        (_(  (_(  (_(\n      /(___((___((___(\n    //(_____(____(____(\n__///____|____|____|_____\n    \\                   /",
	}
	shipMask := []string{
		"     y    y    y\n\n                   w\n                    ww\nyyyyyyyyyyyyyyyyyyyywwwyy\ny                   y",
		"         y    y    y\n\n      w\n    ww\nyywwwyyyyyyyyyyyyyyyyyyyy\n    y                   y",
	}

	dir := rand.Intn(2)
	x := -24.0
	speed := 1.0
	if dir == 1 {
		speed = -1
		x = float64(eng.Width() - 2)
	}

	frames, w, h := ParseFrames([]string{shipShape[dir]})
	colorFrames := ParseColorFrames([]string{shipMask[dir]})
	trans := DetectTransparent(frames)

	eng.AddEntity(&Entity{
		Frames:       frames,
		ColorFrames:  colorFrames,
		Width:        w,
		Height:       h,
		X:            x,
		Y:            0,
		Z:            DepthWaterGap1,
		DX:           speed,
		TransChar:    trans,
		DieOffscreen: true,
		DeathCB:      func(_ *Entity, e *Engine) { randomObject(e) },
		DefStyle:     namedStyle("WHITE"),
	})
}

// ===================== WHALE =====================

func addWhale(_ *Entity, eng *Engine) {
	whaleBody := []string{
		replaceBackticks("        .-----:\n      .'       §.\n,????/       (o) \\\n\\§._/          ,__)"),
		replaceBackticks("    :-----.\n  .'       §.\n / (o)       \\????,\n(__,          \\_.'/"),
	}
	whaleMask := []string{
		"             C C\n           CCCCCCC\n           C  C  C\n        BBBBBBB\n      BB       BB\nB    B       BWB B\nBBBBB          BBBB",
		"   C C\n CCCCCCC\n C  C  C\n    BBBBBBB\n  BB       BB\n B BWB       B    B\nBBBB          BBBBB",
	}

	spoutFrames := []string{
		"\n\n   :",
		"\n   :\n   :",
		"  . .\n  -:-\n   :",
		"  . .\n .-:-.\n   :",
		replaceBackticks("  . .\n'.-:-.§\n'  :  '"),
		"\n .- -.\n;  :  ;",
		"\n\n;     ;",
	}

	dir := rand.Intn(2)
	speed := 1.0
	var x float64
	var spoutAlign int
	if dir == 1 {
		spoutAlign = 1
		speed = -1
		x = float64(eng.Width() - 2)
	} else {
		spoutAlign = 11
		x = -18
	}

	// Build composite animation frames: 5 frames no-spout, 7 frames with spout
	var animFrames []string
	var animMasks []string

	// No spout: 3 blank lines + whale body
	for i := 0; i < 5; i++ {
		animFrames = append(animFrames, "\n\n\n"+whaleBody[dir])
		animMasks = append(animMasks, whaleMask[dir])
	}

	// With spout
	pad := strings.Repeat(" ", spoutAlign)
	for _, sf := range spoutFrames {
		lines := strings.Split(sf, "\n")
		aligned := strings.Join(lines, "\n"+pad)
		animFrames = append(animFrames, aligned+whaleBody[dir])
		animMasks = append(animMasks, whaleMask[dir])
	}

	frames, w, h := ParseFrames(animFrames)
	colorFrames := ParseColorFrames(animMasks)
	trans := DetectTransparent(frames)

	eng.AddEntity(&Entity{
		Frames:       frames,
		ColorFrames:  colorFrames,
		Width:        w,
		Height:       h,
		X:            x,
		Y:            0,
		Z:            DepthWaterGap2,
		DX:           speed,
		AnimSpeed:    1,
		TransChar:    trans,
		DieOffscreen: true,
		DeathCB:      func(_ *Entity, e *Engine) { randomObject(e) },
		DefStyle:     namedStyle("WHITE"),
	})
}

// ===================== MONSTER =====================

func addMonster(_ *Entity, eng *Engine) {
	addNewMonster(eng)
}

func addNewMonster(eng *Engine) {
	monsterFrames := [2][]string{
		{
			"\n         _???_?????????????????????_???_???????_a_a\n       _{.`=`.}_??????_???_??????_{.`=`.}_????{/ ''\\_\n _????{.'  _  '.}????{.`'`.}????{.'  _  '.}??{|  ._oo)\n{ \\??{/  .'?'.  \\}??{/ .-. \\}??{/  .'?'.  \\}?{/  |",
			"\n                      _???_????????????????????_a_a\n  _??????_???_??????_{.`=`.}_??????_???_??????{/ ''\\_\n { \\????{.`'`.}????{.'  _  '.}????{.`'`.}????{|  ._oo)\n  \\ \\??{/ .-. \\}??{/  .'?'.  \\}??{/ .-. \\}???{/  |",
		},
		{
			"\n   a_a_???????_???_?????????????????????_???_\n _/'' \\}????_{.`=`.}_??????_???_??????_{.`=`.}_\n(oo_.  |}??{.'  _  '.}????{.`'`.}????{.'  _  '.}????_\n    |  \\}?{/  .'?'.  \\}??{/ .-. \\}??{/  .'?'.  \\}??/ }",
			"\n   a_a_????????????????????_   _\n _/'' \\}??????_???_??????_{.`=`.}_??????_???_??????_\n(oo_.  |}????{.`'`.}????{.'  _  '.}????{.`'`.}????/ }\n    |  \\}???{/ .-. \\}??{/  .'?'.  \\}??{/ .-. \\}??/ /",
		},
	}
	monsterMask := []string{
		"                                                W W\n\n\n\n",
		"   W W\n\n\n\n",
	}

	dir := rand.Intn(2)
	speed := 2.0
	var x float64
	if dir == 1 {
		speed = -2
		x = float64(eng.Width() - 2)
	} else {
		x = -54
	}

	frames, w, h := ParseFrames(monsterFrames[dir])
	var masks []string
	for range monsterFrames[dir] {
		masks = append(masks, monsterMask[dir])
	}
	colorFrames := ParseColorFrames(masks)
	trans := DetectTransparent(frames)

	eng.AddEntity(&Entity{
		Type:         "monster",
		Frames:       frames,
		ColorFrames:  colorFrames,
		Width:        w,
		Height:       h,
		X:            x,
		Y:            2,
		Z:            DepthWaterGap2,
		DX:           speed,
		AnimSpeed:    0.25,
		TransChar:    trans,
		DieOffscreen: true,
		DeathCB:      func(_ *Entity, e *Engine) { randomObject(e) },
		DefStyle:     namedStyle("GREEN"),
	})
}

// ===================== BIG FISH =====================

func addBigFish(_ *Entity, eng *Engine) {
	if rand.Intn(3) > 1 {
		addBigFish2(eng)
	} else {
		addBigFish1(eng)
	}
}

func addBigFish1(eng *Engine) {
	bigFishShape := []string{
		replaceBackticks(" ______\n§\"\"-.  §§§§§-----.....__\n     §.  .      .       §-.\n       :     .     .       §.\n ,?????:   .    .          _ :\n: §.???:                  (@) §._\n §. §..'     .     =§-.       .__)\n   ;     .        =  ~  :     .-\"\n .' .'§.   .    .  =.-'  §._ .'\n: .'???:               .   .'\n '???.'  .    .     .   .-'\n   .'____....----''.'=.'\n   \"\"?????????????.'.'\n               ''\"'§"),
		replaceBackticks("                           ______\n          __.....-----'''''  .-\"\"'\n       .-'       .      .  .'\n     .'       .     .     :\n    : _          .    .   :?????,\n _.' (@)                  :???.' :\n(__.       .-'=     .     §..' .'\n \"-.     :  ~  =        .     ;\n   §. _.'  §-.=  .    .   .'§. §.\n     §.   .               :???§. :\n       §-.   .     .    .  §.???§\n          §.=§.§§----....____§.\n            §.§.?????????????\"\"\n              '§\"§§"),
	}
	bigFishMask := []string{
		" 111111\n11111  11111111111111111\n     11  2      2       111\n       1     2     2       11\n 1     1   2    2          1 1\n1 11   1                  1W1 111\n 11 1111     2     1111       1111\n   1     2        1  1  1     111\n 11 1111   2    2  1111  111 11\n1 11   1               2   11\n 1   11  2    2     2   111\n   111111111111111111111\n   11             1111\n               11111",
		"                           111111\n          11111111111111111  11111\n       111       2      2  11\n     11       2     2     1\n    1 1          2    2   1     1\n 111 1W1                  1   11 1\n1111       1111     2     1111 11\n 111     1  1  1        2     1\n   11 111  1111  2    2   1111 11\n     11   2               1   11 1\n       111   2     2    2  11   1\n          111111111111111111111\n            1111             11\n              11111",
	}

	dir := rand.Intn(2)
	speed := 3.0
	var x float64
	if dir == 1 {
		x = float64(eng.Width() - 1)
		speed = -3
	} else {
		x = -34
	}
	maxY := 9
	minY := eng.Height() - 15
	if minY <= maxY {
		minY = maxY + 1
	}
	y := rand.Intn(minY-maxY) + maxY

	mask := strings.ReplaceAll(bigFishMask[dir], "4", "W")
	mask = randColor(mask)

	frames, w, h := ParseFrames([]string{bigFishShape[dir]})
	colorFrames := ParseColorFrames([]string{mask})
	trans := DetectTransparent(frames)

	eng.AddEntity(&Entity{
		Frames:       frames,
		ColorFrames:  colorFrames,
		Width:        w,
		Height:       h,
		X:            x,
		Y:            float64(y),
		Z:            DepthShark,
		DX:           speed,
		TransChar:    trans,
		DieOffscreen: true,
		DeathCB:      func(_ *Entity, e *Engine) { randomObject(e) },
		DefStyle:     namedStyle("YELLOW"),
	})
}

func addBigFish2(eng *Engine) {
	bigFishShape := []string{
		replaceBackticks("                _ _ _\n             .='\\  \\ \\§\"=,\n           .'\\  \\ \\ \\ \\ \\ \\\n\\'=._?????/ \\ \\ \\_\\_\\_\\_\\_\\\n\\'=._'.??/\\ \\,-\"§- _ - _ - '-.\n  \\§=._\\|'.\\/-  _ - _ - _ - _- \\\n  ;\"= ._\\=./_ -_ -_ {§\"=_    @ \\\n   ;=\"_-_=- _ -  _ - {\"=_\"-     \\\n   ;_=_--_.,          {_.='   .-/\n  ;.=\"§ / ';\\        _.     _.-§\n  /_.='/ \\/ /;._ _ _{.-;§/\"§\n/._=_.'???'/ / / / /{.= /\n/.=' ??????§'./_/_.=§{_/"),
		replaceBackticks("            _ _ _\n        ,=\"§/ / /'=.\n       / / / / / / /'.\n      /_/_/_/_/_/ / / \\?????_.='/\n   .-' - _ - _ -§\"-,/ /\\??.'_.='/\n  / -_ - _ - _ - _ -\\/.'|/_.=§/\n / @    _=\"§} _- _- _\\.=/_. =\";\n/     -\"_=\"} - _  - _ -=_-_\"=;\n\\-.   '=._}          ,._--_=_;\n §-._     ._        /;' \\ §\"=.;\n     §\"\\§;-.}_ _ _.;\\ \\/ \\'=._\\\n        \\ =.}\\ \\ \\ \\ \\'???'._=_.\\\n         \\_}§=._\\_\\.'§???????'=.\\"),
	}
	bigFishMask := []string{
		"                1 1 1\n             1111 1 11111\n           111 1 1 1 1 1 1\n11111     1 1 1 11111111111\n1111111  11 111112 2 2 2 2 111\n  111111111112 2 2 2 2 2 2 22 1\n  111 1111 12 22 22 11111    W 1\n   11111112 2 2  2 2 111111     1\n   111111111          11111   111\n  11111 11111        11     1111\n  111111 11 1111 1 111111111\n1111111   11 1 1 1 1111 1\n1111       1111111111111",
		"            1 1 1\n        11111 1 1111\n       1 1 1 1 1 1 111\n      11111111111 1 1 1     11111\n   111 2 2 2 2 211111 11  1111111\n  1 22 2 2 2 2 2 2 211111111111\n 1 W    11111 22 22 2111111 111\n1     111111 2 2  2 2 21111111\n111   11111          111111111\n 1111     11        111 1 11111\n     111111111 1 1111 11 111111\n        1 1111 1 1 1 11   1111111\n         1111111111111       1111",
	}

	dir := rand.Intn(2)
	speed := 2.5
	var x float64
	if dir == 1 {
		x = float64(eng.Width() - 1)
		speed = -2.5
	} else {
		x = -33
	}
	maxY := 9
	minY := eng.Height() - 14
	if minY <= maxY {
		minY = maxY + 1
	}
	y := rand.Intn(minY-maxY) + maxY

	mask := randColor(bigFishMask[dir])

	frames, w, h := ParseFrames([]string{bigFishShape[dir]})
	colorFrames := ParseColorFrames([]string{mask})
	trans := DetectTransparent(frames)

	eng.AddEntity(&Entity{
		Frames:       frames,
		ColorFrames:  colorFrames,
		Width:        w,
		Height:       h,
		X:            x,
		Y:            float64(y),
		Z:            DepthShark,
		DX:           speed,
		TransChar:    trans,
		DieOffscreen: true,
		DeathCB:      func(_ *Entity, e *Engine) { randomObject(e) },
		DefStyle:     namedStyle("YELLOW"),
	})
}

// ===================== RANDOM OBJECT SPAWNER =====================

func randomObject(eng *Engine) {
	spawners := []func(*Entity, *Engine){
		addShip,
		addWhale,
		addMonster,
		addBigFish,
		addShark,
	}
	fn := spawners[rand.Intn(len(spawners))]
	fn(nil, eng)
}
