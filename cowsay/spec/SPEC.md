# cowsay Specification

## 1. CLI Flags and exact semantics
- **Mood Flags (with precedence if combined):**
  - `-b` (Borg): eyes `==`, tongue `  `
  - `-d` (Dead): eyes `xx`, tongue `U `
  - `-g` (Greedy): eyes `$$`, tongue `  `
  - `-p` (Paranoia): eyes `@@`, tongue `  `
  - `-s` (Stoned): eyes `**`, tongue `U `
  - `-t` (Tired): eyes `--`, tongue `  `
  - `-w` (Wired): eyes `OO`, tongue `  `
  - `-y` (Youthful): eyes `..`, tongue `  `
  - *Precedence:* Mood flags are resolved strictly in alphabetical order of their defining character (`b, d, g, p, s, t, w, y`). The first matching flag forces the eyes and tongue.
- **Custom Overrides:**
  - `-e` overrides the default eyes (`oo`).
  - `-T` overrides the default tongue (`  `).
  - *Resolution Order:* Mood flags completely override custom `-e` and `-T` flags. If any mood flag is present, its defined eyes and tongue take precedence over anything passed via `-e` or `-T`.
- **Cowfile Selection (`-f`):**
  - `-f <name>` selects a cow template. Defaults to `default`.
- **Random Cow (`-r`):**
  - Randomly selects a cowfile from the bundled cows list.
- **List Mode (`-l`):**
  - Prints all available bundled cow files separated exactly by two spaces (`  `), in alphabetical order, then exits.
- **Wrap-Width (`-W`):**
  - Defaults to `40`.
  - Determines the maximum width of the speech bubble text.
  - No environment variables are consulted for this width.
- **No-Wrap (`-n`):**
  - Disables word-wrapping entirely.
  - Structurally, it passes `null` instead of a number for the width, causing lines to span indefinitely until an explicit newline is encountered.
- **Say vs Think Mode:**
  - Determined by the `--think` flag, or by invoking the program as `cowthink` (based on the executable name).

## 2. Message Source Precedence
- **CLI Positional Args vs Stdin:** Positional CLI arguments (any unparsed trailing strings) take strict precedence. 
- **Conflict Resolution:** If both positional args and piped stdin are provided, positional arguments are joined with spaces to form the message, and stdin is entirely ignored.

## 3. Default Invocation
- **Conditions:** Program is run with no message argument and no piped stdin (e.g. empty TTY).
- **Placeholder Message:** It does NOT print a placeholder message.
- **Output:** It prints the exact CLI help text showing usage and options.
- **Borders:** There are NO top or bottom border characters printed because no bubble is drawn.
- **Exit Code:** Exits with code `0`.

## 4. Word-Wrap Algorithm
- **Overlong Single Words:** A single word that exceeds the column limit is hard-wrapped exactly at the column limit (broken into multiple chunks).
- **Existing Newlines:** Any existing newlines in the input string are preserved and force a line break, restarting the column count for the next line.
- **Whitespace Collapsing:** Whitespace is collapsed around word boundaries when wrapping occurs.

## 5. Bubble-Drawing Algorithm
- **Say Mode (`cowsay`):**
  - **Single-line borders:** Left: `<`, Right: `>`.
  - **Multi-line borders:** 
    - First line: Left: `/`, Right: `\`.
    - Middle lines: Left: `|`, Right: `|`.
    - Last line: Left: `\`, Right: `/`.
  - **Connector:** `\`
- **Think Mode (`cowthink`):**
  - **Single-line & Multi-line borders:** Left: `(`, Right: `)` on all lines.
  - **Connector:** `o`
- **Common Border Rules:**
  - **Top Border:** A space followed by underscores (`_`). The number of underscores is equal to the longest wrapped line length plus 2.
  - **Bottom Border:** A space followed by dashes (`-`). The number of dashes is equal to the longest wrapped line length plus 2.
- **Zero-Length Message (`""`):**
  - The top border is a space and two underscores (` __`).
  - The bottom border is a space and two dashes (` --`).
  - **No side borders** are drawn at all.

## 6. Error Handling
- **Unknown Cowfile:** Prints `Error: ENOENT: no such file or directory, open '<path>'` (plus a stack trace). Exit code `1`.
- **Empty Message Passed (`""`):** Prints a 0-width empty bubble (` __` and ` --` with no side borders) and the selected cow. Exit code `0`.
- **Malformed Flags:** Unknown flags (e.g., `--invalid-flag`) are parsed as boolean or string options by the CLI parser, effectively consuming the following positional arguments. If it absorbs the intended message, the program behaves as if no message was provided, resulting in printing the help text (exit code 0).

## 7. Nondeterminism
- **Random Cow (`-r`):** Seeded by the internal JavaScript engine state (`Math.random()`). It is not explicitly seeded and thus is NOT reproducible across identical runs.

## 8. Cow Template Format
- **Placeholder Tokens:**
  - `$thoughts`: Replaced by the connector character (`\` for say, `o` for think).
  - `$eyes`: Replaced by the current eye string (e.g., `oo`, `==`).
  - `$tongue`: Replaced by the current tongue string (e.g., `  `, `U `).
- **Bundled Cowfiles Reference Data:**

### ackbar
```perl
# Admiral Ackbar
#
# based on 'ack --bar' from http://beyondgrep.com/
$the_cow = <<EOC;
         $thoughts
          $thoughts
                      ?IIIIIII7II?????+
                   ~III777II777I?+==++==+:
                  ???I7I???I7II++=====++===
                 ??+??????????+===~~=+++??==+
                ??+??II??????+==~=~~=+++++==++
               I+?????????+?+====~=~==+==++?==?
              ?????II?????+++++=======?===~~~~==
            ,?????II????????++++====~===::~~~~:~
            I?I??II?+++??+?+++==~~~~:~:~:,:,,:::~
           I??????+==+???++++=~~:~:~:,:::,:,,,,,::
          +I?++++=+=+????+++=~~:~~:::,,,,::,,,:,:
          I??+?+====+???+++===~~::::,::,:,,,,,,::
         I????=~===++?+=+=~==~:~~:,,,,,,,.,,,,,:~
        =??+?=~~~~??+?+===~~==,==~~~~,,,,..,,,.:=
        II++==~~=++++++=~~=~,~+=?+?=I?++=..,.,,:
     IIII?+?=====+~+++~=~~~:::=~+~===:,,,,,.,.::
    I?=?I+??+=~=~?I?=+=~~~::,~~=~::~=::,,,,,,::
    ?+I??++=++~,::+++~~~:::,,=~~=,~,..,::.:
    ++=+?++~=:~::I+,~=:~,:,,,,:~~......::~,,,
     ~=~=:.++~:,.,~=::::.,,:,.:~,:=...==~,::
     =~?++??+=~~,.:?~.:,:,,,.,::,,~:=~=::,~
     ++~~:~===~:~,.~::,~=~.:,..,:,,:==:.,:7
     ~~,::...:=:,::+:~:.,~,...,.,,,,::~,,::~=
      =~===+=~~,.::,,,:::,..,,,,,,,,,,,:,..,=+?
      ~=~=~::~~~::,.,,,~:.+,..,,,,..,,,,...,+I?
      ~==~:~~:~~,~=~~:,:~,:,,,,,,....,,,..+?I?I
      ~=~=+,:~:=,:~~~~~~::::,.,,.,,.,,,..~+????I
      ~=~==~=:~~:,~~~~~:::,::,.,,,..,,,I77I?+??II
      +I7:::~~=~:,::~~~~.=.,~,,,,...,~7III?+??II7
     777?+~:=~=~~:,::~~:::.,,,,,,,,,777II??I777777
     777I==:=~::~~~~::~:::,:,:~:::,777I???777777777
    7777+,~===~:~:~~~~:::,.~:=,,:777II???77777777777=?
    777I~,~~~=~::~:,:,,,:=~~,,:7777I???I7777777777+=++
  I7777I,,:,.==::::,:,,,,::::7777I+??I77777777777??I7I7,
 ,77777I::,..~~:,,,,,,.,:~I7777I+??I777777777777?I7777777,
 77777777,...~~:,,,,,.,77777I7???II777777777777+?7777777777
77777777777:,~~~,,=7777777I???II777777777777777+77777777777
77777777777777777777777I+7?7II77777777777777777+777777777777
EOC


```

### aperture-blank
```perl
# Aperture Science logo, without the text inside
# via http://pastebin.com/1AZwKrKp 
$the_cow = <<EOC;
    $thoughts
     $thoughts
              .,-:;//;:=,
          . :H\@\@\@MM\@M#H/.,+%;,
       ,/X+ +M\@\@M\@MM%=,-%HMMM\@X/,
     -+\@MM; \$M\@\@MH+-,;XMMMM\@MMMM\@+-
    ;\@M\@\@M- XM\@X;. -+XXXXXHHH\@M\@M#\@/.
  ,%MM\@\@MH ,\@%=            .---=-=:=,.
  =\@#\@\@\@MX .,              -%HX\$\$%%%+;
 =-./\@M\@M\$                  .;\@MMMM\@MM:
 X\@/ -\$MM/                    .+MM\@\@\@M\$
,\@M\@H: :\@:                    . =X#\@\@\@\@-
,\@\@\@MMX, .                    /H- ;\@M\@M=
.H\@\@\@\@M\@+,                    %MM+..%#\$.
 /MMMM\@MMH/.                  XM\@MH; =;
  /%+%\$XHH\@\$=              , .H\@\@\@\@MX,
   .=--------.           -%H.,\@\@\@\@\@MX,
   .%MM\@\@\@HHHXX\$\$\$%+- .:\$MMX =M\@\@MM%.
     =XMMM\@MM\@MM#H;,-+HMM\@M+ /MMMX=
       =%\@M\@M#\@\$-.=\$\@MM\@\@\@M; %M%=
         ,:+\$+-,/H#MMMMMMM\@= =,
               =++%%%%+/:-.
EOC

```

### aperture
```perl
# Aperture Science logo
# via http://pastebin.com/1AZwKrKp 
$the_cow = <<EOC;
    $thoughts
     $thoughts
              .,-:;//;:=,
          . :H\@\@\@MM\@M#H/.,+%;,
       ,/X+ +M\@\@M\@MM%=,-%HMMM\@X/,
     -+\@MM; \$M\@\@MH+-,;XMMMM\@MMMM\@+-
    ;\@M\@\@M- XM\@X;. -+XXXXXHHH\@M\@M#\@/.
  ,%MM\@\@MH ,\@%=            .---=-=:=,.
  =\@#\@\@\@MX .,      WE      -%HX\$\$%%%+;
 =-./\@M\@M\$         DO       .;\@MMMM\@MM:
 X\@/ -\$MM/        WHAT        .+MM\@\@\@M\$
,\@M\@H: :\@:         WE         . =X#\@\@\@\@-
,\@\@\@MMX, .        MUST        /H- ;\@M\@M=
.H\@\@\@\@M\@+,      BECAUSE       %MM+..%#\$.
 /MMMM\@MMH/.       WE         XM\@MH; =;
  /%+%\$XHH\@\$=     CAN      , .H\@\@\@\@MX,
   .=--------.           -%H.,\@\@\@\@\@MX,
   .%MM\@\@\@HHHXX\$\$\$%+- .:\$MMX =M\@\@MM%.
     =XMMM\@MM\@MM#H;,-+HMM\@M+ /MMMX=
       =%\@M\@M#\@\$-.=\$\@MM\@\@\@M; %M%=
         ,:+\$+-,/H#MMMMMMM\@= =,
               =++%%%%+/:-.
EOC

```

### armadillo
```perl
# armadillo
#
# based on http://ascii.co.uk/art/armadillo
$the_cow = <<EOC;
         $thoughts
          $thoughts
               ,.-----__
            ,:::://///,:::-.
           /:''/////// ``:::`;/|/
          /'   ||||||     :://'`\\
        .' ,   ||||||     `/(  e \\
  -===~__-'\\__X_`````\\_____/~`-._ `.
              ~~        ~~       `~-'
EOC


```

### atat
```perl
# ATAT
# from http://www.asciiworld.com/-Robots,24-.html (accessed 4/30/2014)
$the_cow = <<EOC;
  $thoughts                         ________
   $thoughts                    _.-Y  |  |  Y-.,_
    $thoughts                .-"   |  |  |  ||   "~-.      
          _____     |""[]"|" !""! "|"=="" "I      
       .-"{-. "I----]_   :|------..| []  __L      
      P-=}=(r\_I]_[L__] _l|______l |..  |___I     
      ^-=\[_c=-'  ~j______[________]_L______L]    
                    [_L--.\_==I|I==/.--.j_I_/     
                      j)==(["-----`])==((_]       
                       I--I"~~"""~~"I--I          
                       |[]|         |[]|          
                       j__l         j__l          
                       |!!|         |!!|          
                       |..|         |..|         
                       )[](         )[](          
                       ]--[         ]--[          
                       [L_]         [L_]          
                      /|..|\       /|..|\         
                     '={--}=`     '={--}=`        
                    .-^-r--^-.   .-^-r--^-.       
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Modified ATAT from Row  (the Ascii-Wizard of Oz)
EOC

```

### atom
```perl
# atom
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
       $thoughts
        $thoughts
                  =/;;/-
                 +:    //
                /;      /;
               -X        H.
 .//;;;:;;-,   X=        :+   .-;:=;:;%;.
 M-       ,=;;;#:,      ,:#;;:=,       ,\@
 :%           :%.=/++++/=.\$=           %=
  ,%;         %/:+/;,,/++:+/         ;+.
    ,+/.    ,;\@+,        ,%H;,    ,/+,
       ;+;;/= \@.  .H##X   -X :///+;
       ;+=;;;.\@,  .XM\@\$.  =X.//;=%/.
    ,;:      :\@%=        =\$H:     .+%-
  ,%=         %;-///==///-//         =%,
 ;+           :%-;;;:;;;;-X-           +:
 \@-      .-;;;;M-        =M/;;;-.      -X
  :;;::;;-.    %-        :+    ,-;;-;:==
               ,X        H.
                ;/      %=
                 //    +;
                  ,////,

EOC

```

### awesome-face
```perl
# awesome face
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
    $thoughts
     $thoughts
                  \#[/[#:xxxxxx:#[/[\\x
             [/\\ &3N            W3& \\/[x
          [[x\@W                      W\@x[[\\
        /#&N                             N_#
      /#\@                                  \@#/x
    [/ NH_  ^\@W               Nd_  ^\@p      N /#
   [[d\@#_ zz\@[/x3           3x:d9zz \\/#_N     d[[
  /[3^[JMMMJ/////&         ^#NMMMMM ////#W     H[[
 [/\@p/NMMMML\@#[:^/3       d/JMMMMMMEx[# x\\      &/#
 /x &/LMMMMMMMMMM[_       x:MMMMMMMMMMMM /p      :/
[/d d/ELLLLLLLLLD/&        \#LLLLLLLLLLLL3/N      d/[
//N   xxxxxxxxxxxxN       Wxxxxxxxxxxxxxx_       W//
/[                                                //
//N   p333333333333333333333333333333333p        W//
[/d   _^/#\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\/H       \@/[
 /:     \\#                              [x       :/
 [/\@    d/x                             \#:      &/#
  [[H    ^[x                            [      H[[
   [[d    _[x            &Hppp3d_      \#\\N    \@[[
    [/ N   d#\\        &NzDDDDDDDDJp^ x[xN   N /#
      /#&   N [:     pDDDDDDDDDDDDJ&#:H    &#/
       :/#_W  W^##x 3DDDDDDDDDJN&:\\^p   W_#/
          [[x&W  p& xx ^^^^ x:x \@W   W&x/[
             [/# &HW   WWWWN    WH& \#/[
                 [/[#\\xxxxxx\\#[/[\\x^\@
EOC

```

### banana
```perl
# Banana 
#  http://www.ascii-art.de/ascii/ab/banana.txt
$the_cow = <<EOC;
       $thoughts
        $thoughts

     ".           ,#  
     \\ `-._____,-'=/
  ____`._ ----- _,'_____PhS
         `-----'
EOC

```

### bearface
```perl
##
## acsii picture from http://www.ascii-art.de/ascii/ab/bear.txt
##
$the_cow = <<EOC;
 $thoughts
  $thoughts
     .--.              .--.
    : (\\ ". _......_ ." /) :
     '.    `        `    .'
      /'   _        _   `\\
     /     $eye}      {$eye     \\
    |       /      \\       |
    |     /'        `\\     |
     \\   | .  .==.  . |   /
      '._ \\.' \\__/ './ _.'
      /  ``'._-''-_.'``  \\
EOC

```

### beavis.zen
```perl
##
## Beavis, with Zen philosophy removed.
##
$the_cow = <<EOC;
   $thoughts         __------~~-,
    $thoughts      ,'            ,
          /               \\
         /                :
        |                  '
        |                  |
        |                  |
         |   _--           |
         _| =-.     .-.   ||
         $eye|/$eye/       _.   |
         /  ~          \\ |
       (____\@)  ___~    |
          |_===~~~.`    |
       _______.--~     |
       \\________       |
                \\      |
              __/-___-- -__
             /            _ \\
EOC

```

### bees
```perl
# Bees/beehive
#  http://www.asciiworld.com/-Bees-.html
$the_cow = <<EOC;
          $thoughts
           $thoughts


      ^^      .-=-=-=-.  ^^
  ^^        (`-=-=-=-=-`)         ^^
          (`-=-=-=-=-=-=-`)  ^^         ^^
    ^^   (`-=-=-=-=-=-=-=-`)   ^^                            ^^
        ( `-=-=-=-(@)-=-=-` )      ^^
        (`-=-=-=-=-=-=-=-=-`)  ^^
        (`-=-=-=-=-=-=-=-=-`)              ^^
        (`-=-=-=-=-=-=-=-=-`)                      ^^
        (`-=-=-=-=-=-=-=-=-`)  ^^
         (`-=-=-=-=-=-=-=-`)          ^^
          (`-=-=-=-=-=-=-`)  ^^                 ^^
      jgs   (`-=-=-=-=-`)
             `-=-=-=-=-`
EOC

```

### bill-the-cat
```perl
# Bill the Cat
#
# Based on 'ack --th[pt]+t+'
#  from http://beyondgrep.com/ack-2.14-single-file
$the_cow = <<EOC;
 $thoughts
  $thoughts
 _   /|
 \\'o.O'
 =(___)=
    U
EOC

```

### biohazard
```perl
# biohazard symbol
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
     $thoughts
      $thoughts
              =+\$HM####\@H%;,
           /H###############M\$,
           ,\@################+
            .H##############+
              X############/
               \$##########/
                %########/
                 /X/;;+X/
 
                  -XHHX-
                 ,######,
 \#############X  .M####M.  X#############
 \##############-   -//-   -##############
 X##############%,      ,+##############X
 -##############X        X##############-
  %############%          %############%
   %##########;            ;##########%
    ;#######M=              =M#######;
     .+M###\@,                ,\@###M+.
        :XH.                  .HX:

EOC

```

### bishop
```perl
# Bishop (Chess piece)
#
# from http://www.chessvariants.org/d.pieces/ascii.html
#   by David Moeser
#
$the_cow = <<EOC;
 $thoughts
  $thoughts
    <>_
  (\\)  )
   \\__/
  (____)
   |  |
   |__|
  /____\\
 (______)
EOC

```

### black-mesa
```perl
# Black Mesa logo
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
    $thoughts
     $thoughts
           .-;+\$XHHHHHHX\$+;-.
        ,;X\@\@X%/;=----=:/%X\@\@X/,
      =\$\@\@%=.              .=+H\@X:
    -XMX:                      =XMX=
   /\@\@:                          =H\@+
  %\@X,                            .\$\@\$
 +\@X.                               \$\@%
-\@\@,                                .\@\@=
%\@%                                  +\@\$
H\@:                                  :\@H
H\@:         :HHHHHHHHHHHHHHHHHHX,    =\@H
%\@%         ;\@M\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@H-   +\@\$
=\@\@,        :\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@= .\@\@:
 +\@X        :\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@M\@\@\@\@\@\@:%\@%
  \$\@\$,      ;\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@M\@\@\@\@\@\@\$.
   +\@\@HHHHHHH\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@+
    =X\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@X=
      :\$\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@M\@\@\@\@\$:
        ,;\$\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@X/-
           .-;+\$XXHHHHHX\$+;-.
EOC

```

### bong
```perl
##
## A cow with a bong, from lars@csua.berkeley.edu
##
$the_cow = <<EOC;
         $thoughts
          $thoughts
            ^__^ 
    _______/($eyes)
/\\/(       /(__)
   | W----|| |~|
   ||     || |~|  ~~
             |~|  ~
             |_| o
             |#|/
            _+#+_
EOC

```

### box
```perl
# Box
$the_cow = <<EOC;
     $thoughts
      $thoughts
         __________________
        /\\  ______________ \\
       /::\\ \\ZZZZZZZZZZZZ/\\ \\
      /:/\\.\\ \\        /:/\\:\\ \\
     /:/Z/\\:\\ \\      /:/Z/\\:\\ \\
    /:/Z/__\\:\\ \\____/:/Z/  \\:\\ \\
   /:/Z/____\\:\\ \\___\\/Z/    \\:\\ \\
   \\:\\ \\ZZZZZ\\:\\ \\ZZ/\\ \\     \\:\\ \\
    \\:\\ \\     \\:\\ \\ \\:\\ \\     \\:\\ \\
     \\:\\ \\     \\:\\ \\_\\;\\_\\_____\\;\\ \\
      \\:\\ \\     \\:\\_________________\\
       \\:\\ \\    /:/ZZZZZZZZZZZZZZZZZ/
        \\:\\ \\  /:/Z/    \\:\\ \\  /:/Z/
         \\:\\ \\/:/Z/      \\:\\ \\/:/Z/
          \\:\\/:/Z/________\\;\\/:/Z/
           \\::/Z/_______itz__\\/Z/
            \\/ZZZZZZZZZZZZZZZZZ/
EOC

```

### broken-heart
```perl
# broken heart
# via http://pastebin.com/1AZwKrKp
# TODO: replace "thoughts" with "feelings"
$the_cow = <<EOC;
     $thoughts
      $thoughts
                          .,---.
                        ,/XM#MMMX;,
                      -%##########M%,
                     -\@######%  \$###\@=
      .,--,         -H#######\$   \$###M:
   ,;\$M###MMX;     .;##########\$;HM###X=
 ,/\@##########H=      ;################+
-+#############M/,      %##############+
%M###############=      /##############:
H################      .M#############;.
\@###############M      ,\@###########M:.
X################,      -\$=X#######\@:
/\@##################%-     +######\$-
.;##################X     .X#####+,
 .;H################/     -X####+.
   ,;X##############,       .MM/
      ,:+\$H\@M#######M#\$-    .\$\$=
           .,-=;+\$\@###X:    ;/=.
                  .,/X\$;   .::,
                      .,    ..
EOC

```

### bud-frogs
```perl
##
## The Budweiser frogs
##
$the_cow = <<EOC;
     $thoughts
      $thoughts
          oO)-.                       .-(Oo
         /__  _\\                     /_  __\\
         \\  \\(  |     ()~()         |  )/  /
          \\__|\\ |    (-___-)        | /|__/
          '  '--'    ==`-'==        '--'  '
EOC

```

### bunny
```perl
##
## A cute little wabbit
##
$the_cow = <<EOC;
  $thoughts
   $thoughts   \\
        \\ /\\
        ( )
      .( o ).
EOC

```

### C3PO
```perl
# C3PO
#
# adapted from 'telnet -e x towel.blinkenlights.nl'
$the_cow = <<EOC;
   $thoughts
    $thoughts
       /~\\
      |oo )
      _\\=/_
     /     \\
    //|/.\\|\\\\
   ||  \\_/  ||
   || |\\ /| ||
    \# \\_ _/  \#
      | | |
      | | |
      []|[]
      | | |
     /_]_[_\\
EOC

```

### cake-with-candles
```perl
# cake with candles
# via http://chris.com/ascii/index.php?art=events/birthday
$the_cow = <<EOC;
     $thoughts
      $thoughts
       $thoughts
                                    (
                       (
               )                    )             (
                       )           (o)    )
               (      (o)    )     ,|,            )
              (o)     ,|,          |~\\    (      (o)
              ,|,     |~\\    (     \\ |   (o)     ,|,
              \\~|     \\ |   (o)    |`\\   ,|,     |~\\
              |`\\     |`\\\@\@\@,|,\@\@\@\@\\ |\@\@\@\\~|     \\ |
              \\ | o\@\@\@\\ |\@\@\@\\~|\@\@\@\@|`\\\@\@\@|`\\\@\@\@o |`\\
             o|`\\\@\@\@\@\@|`\\\@\@\@|`\\\@\@\@\@\\ |\@\@\@\\ |\@\@\@\@\@\\ |o
           o\@\@\\ |\@\@\@\@\@\\ |\@\@\@\\ |\@\@\@\@\@\@\@\@\@\@|`\\\@\@\@\@\@|`\\\@\@o
          \@\@\@\@|`\\\@\@\@\@\@\@\@\@\@\@\@|`\\\@\@\@\@\@\@\@\@\@\@\\ |\@\@\@\@\@\\ |\@\@\@\@
          p\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\\ |\@\@\@\@\@\@\@\@\@\@|`\\\@\@\@\@\@\@\@\@\@\@\@q
          \@\@o\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@|`\\\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@o\@\@
          \@:\@\@\@o\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@o\@\@::\@
          ::\@\@::\@\@o\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@\@o\@\@:\@\@::\@
          ::\@\@::\@\@\@\@::oo\@\@\@\@oo\@\@\@\@\@ooo\@\@\@\@\@o:::\@\@\@::::::
          %::::::\@::::::\@\@\@\@:::\@\@\@:::::\@\@\@\@:::::\@\@:::::%
          %%::::::::::::\@\@::::::\@:::::::\@\@::::::::::::%%
          ::%%%::::::::::\@::::::::::::::\@::::::::::%%%::
        .#::%::%%%%%%:::::::::::::::::::::::::%%%%%::%::#.
      .###::::::%%:::%:%%%%%%%%%%%%%%%%%%%%%:%:::%%:::::###.
    .#####::::::%:::::%%::::::%%%%:::::%%::::%::::::::::#####.
   .######`:::::::::::%:::::::%:::::::::%::::%:::::::::\'######.
   .#########``::::::::::::::::::::::::::::::::::::\'\'#########.
   `.#############```::::::::::::::::::::::::\'\'\'#############.\'
    `.######################################################.\'
      ` .###########,._.,,,. \#######<_\\##################. \'
         ` .#######,;:      `,/____,__`\\_____,_________,_____
            `  .###;;;`.   _,;>-,------,,--------,----------\'
                `  `,;\' ~~~ ,\'\\######_/\'#######  .  \'
                    \'\'~`\'\'\'\'    -  .\'/;  -    \'       -Catalyst
EOC

```

### cake
```perl
# Cake, from Portal 
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
    $thoughts
     $thoughts
            ,:/+/-
            /M/              .,-=;//;-
       .:/= ;MH/,    ,=/+%\$XH@MM#@:
      -\$##@+\$###@H@MMM#######H:.    -/H#
 .,H@H@ X######@ -H#####@+-     -+H###@X
  .,@##H;      +XM##M/,     =%@###@X;-
X%-  :M##########$.    .:%M###@%:
M##H,   +H@@@$/-.  ,;\$M###@%,          -
M####M=,,---,.-%%H####M\$:          ,+@##
@##################@/.         :%H##@\$-
M###############H,         ;HM##M\$=
\#################.    .=\$M##M\$=
\#################H..;XM##M\$=          .:+
M###################@%=           =+@MH%
@################M/.          =+H#X%=
=+M##############M,       -/X#X+;.
  .;XM##########H=    ,/X#H+:,
     .=+HM######M+/+HM@+=.
         ,:/%XM####H/.
              ,.:=-.
EOC

```

### cat
```perl
# Cat
#
# used https://github.com/paulkaefer/flipFile.py
#  python flipFile.py cat " "
# and 
#  cat cat_flipped | sed 's/\\/\\\\/g' > cat.cow
#
$the_cow = <<EOC;
  $thoughts
   $thoughts                       _
                          / )      
                         / /       
      //|                \\ \\       
   .-`^ \\   .-`````-.     \\ \\      
 o` {|}  \\_/         \\    / /      
 '--,  _ //   .---.   \\  / /       
   ^^^` )/  ,/     \\   \\/ /        
        (  /)      /\\/   /         
        / / (     / (   /          
    ___/ /) (  __/ __\\ (           
   (((__)((__)((__(((___)          
EOC


```

### cat2
```perl
#
#	Cat picture by Joan Stark
#	Transformed into cowfile by Myroslav Golub
#
$the_cow = <<EOC;
       $thoughts  
        $thoughts
         $thoughts
          $thoughts
          |\\___/|
         =) $eyeY$eye (=            
          \\  ^  /
           )=*=(       
          /     \\
          |     |
         /| | | |\\
         \\| | |_|/\\
         //_// ___/
             \\_) 
EOC

```

### catfence
```perl
#
#	Cat picture by Joan Stark
#	Transformed into cowfile by Myroslav Golub
#
$the_cow = <<EOC;
       $thoughts     *     ,MMM8&&&.            *
                  MMMM88&&&&&    .
        $thoughts        MMMM88&&&&&&&
     *           MMM88&&&&&&&&
         $thoughts       MMM88&&&&&&&&
                 'MMM88&&&&&&'
          $thoughts        'MMM8&&&'      *
          |\\___/|
         =) $eyeY$eye (=            .              '
          \\  ^  /
           )=*=(       *
          /     \\
          |     |
         /| | | |\\
         \\| | |_|/\\
  _/\\_/\\_//_// ___/\\_/\\_/\\_/\\_/\\_/\\_/\\_/\\_/\\_
  |  |  |  | \\_) |  |  |  |  |  |  |  |  |  |
  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |     
  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |
  |  |  |  |  |  |  |  |  |  |  |  |  |  |  |

EOC

```

### charizardvice
```perl
$the_cow = <<"EOC";
                        $thoughts
                         $thoughts     ___.
                          $thoughts    L._, \\
               _.,         $thoughts   <  <\\                _
             ,' '           $thoughts  `.   | \\            ( `
          ../, `.            $thoughts  |    .\\`.           \\ \\_
         ,' ,..  .           _.,'    ||\\l            )  '".
        , ,'   \\           ,'.-.`-._,'  |           .  _._`.
      ,' /      \\ \\        `' ' `--/   | \\          / /   ..\\
    .'  /        \\ .         |\\__ - _ ,'` `        / /     `.`.
    |  '          ..         `-...-"  |  `-'      / /        . `.
    | /           |L__           |    |          / /          `. `.
   , /            .   .          |    |         / /             ` `
  / /          ,. ,`._ `-_       |    |  _   ,-' /               ` \\
 / .           \\"`_/. `-_ \\_,.  ,'    +-' `-'  _,        ..,-.    \\`.
  '         .-f    ,'   `    '.       \\__.---'     _   .'   '     \\ \\
' /          `.'    l     .' /          \\..      ,_|/   `.  ,'`     L`
|'      _.-""` `.    \\ _,'  `            \\ `.___`.'"`-.  , |   |    | \\
||    ,'      `. `.   '       _,...._        `  |    `/ '  |   '     .|
||  ,'          `. ;.,.---' ,'       `.   `.. `-'  .-' /_ .'    ;_   ||
|| '              V      / /           `   | `   ,'   ,' '.    !  `. ||
||/            _,-------7 '              . |  `-'    l         /    `||
 |          ,' .-   ,' ||               | .-.        `.      .'     ||
 `'        ,'    `".'    |               |    `.        '. -.'       `'
          /      ,'      |               |,'    \\-.._,.'/'
          .     /        .               .       \\    .''
        .`.    |         `.             /         :_,'.'
          \\ `...\\   _     ,'-.        .'         /_.-'
           `-.__ `,  `'   .  _.>----''.  _  __  /
                .'        /"'          |  "'   '_
               /_|.-'\\ ,".             '.'`__'-( \\
                 / ,"'"\\,'               `/  `-.|" m
EOC

```

### charlie
```perl
##
## KMB is God.
##
$the_cow = <<EOC;
  $thoughts
   $thoughts
    $thoughts     ,, ＿
        ／      ｀､
       /   (_ﾉL_） ヽ
      /   ´・ ・｀  l
    （l      し     l）
      l     ＿＿    l
      >  ､ _      ィ
    ／        ￣    ヽ
   /  |              iヽ
   |＼|              |/|
   |  ||/＼／＼／＼/ | |
EOC

```

### cheese
```perl
##
## The cheese from milk & cheese
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
      _____   _________
     /     \\_/         |
    |                 ||
    |                 ||
   |    ###\\  /###   | |
   |     $eye  \\/  $eye    | |
  /|                 | |
 / |        <        |\\ \\
| /|                 | | |
| |     \\_______/   |  | |
| |        $tongue       | / /
/||                 /|||
   ----------------|
        | |    | |
        ***    ***
       /___\\  /___\\
EOC

```

### chessmen
```perl
# Chessmen Lineup
#
# based on ASCII chess pieces from http://www.chessvariants.org/d.pieces/ascii.html
#
# used https://github.com/paulkaefer/connectFiles.py
#   to "glue" the pieces together into one file
$the_cow = <<EOC;
    $thoughts
     $thoughts 
      $thoughts
       $thoughts
                                           .::.                      
                                           _::_                      
                                 ()      _/____\\_                    
                               <~~~~>    \\      /                    
                       <>_      \\__/      \\____/      <>_            
           __/"""\\   (\\)  )    (____)     (____)    (\\)  )   __/"""\\ 
  WWWWWW  ]___ 0  }   \\__/      |  |       |  |      \\__/   ]___ 0  }  WWWWWW
   |  |       /   }  (____)     |  |       |__|     (____)      /   }   |  |
   |  |     /~    }   |  |      |__|      /    \\     |  |     /~    }   |  |
   |__|     \\____/    |__|     /____\\    (______)    |__|     \\____/    |__|
  /____\\    /____\\   /____\\   (______)  (________)  /____\\    /____\\   /____\\
 (______)  (______) (______) (________) /________\\ (______)  (______) (______)

    __        __       __        __         __        __        __       __
   (  )      (  )     (  )      (  )       (  )      (  )      (  )     (  )
    ||        ||       ||        ||         ||        ||        ||       ||
   /__\\      /__\\     /__\\      /__\\       /__\\      /__\\      /__\\     /__\\
  (____)    (____)   (____)    (____)     (____)    (____)    (____)   (____)
EOC

```

### chito
```perl
#
# ちーちゃん
#
$the_cow = <<EOC;
   $thoughts
    $thoughts
                -一     一-
        ／                       ＼
       /             ________
      /     -~                     ミ､
      レ'     _  一ｧiァ ￢}￣Tii一- _  ＼
    ／    --::|::::/斗士  /   |[_Vい＿＞」
  ／ イ「::::|:::Y/  ｲ::ハ      ｨ-ﾐヽい
  ＜___｜:::へ|::|{ 乂-夕     {::ｄﾘ|い
        ＼八 |::｜             `''   ﾊ|
    ＿ --＼ヽ|::|                  .ｲ ﾘ
  ／------.ゝ|:ﾄ|        -       ィ:|
  ＼        ＞ミ|`ヽ!ﾆ  T  ﾌ￣.≧｜:/
     ∨         |::\/ }-/く＼   /｜/ 
EOC

```

### claw-arm
```perl
# claw arm
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
  $thoughts
   $thoughts
       X MM X
       X MM X
       X MM X
       X MM X
       + HX +
     ,=\$\$XX%/-
   =X#########\@%-
  ;##############=
 -###############M,
 ;##\@\@\@######M\@###=
 .+:;+:=H##\$=:/:;H.
 - +###- \## :###,,;
 +\@:/%;-H##H==/::H;
  /#\@/-=+\$\$%::+H#\$
  \$#%-,      ,.:##-
 -\@/            =X%.
 %H=             -\$;
  =HH,         .%M;
   /MM/       :\@M/.
    .:XX,   -\$H:.
EOC

```

### clippy
```perl
# Clippy
#
# from http://www.reddit.com/r/commandline/comments/2lb5ij/what_is_your_favorite_ascii_art/cltg01p
#
$the_cow = <<EOC;
 $thoughts
  $thoughts
     __ 
    /  \\  
    |  |
    @  @
    |  |
    || |/ 
    || || 
    |\\_/|
    \\___/
EOC


```

### companion-cube
```perl
# Companion Cube from Portal
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
     $thoughts
      $thoughts

 +\@##########M/             :\@#########\@/
 \##############\$;H#######\@;+#############
 \###############M########################
 \##############X,-/++/+%+/,%#############
 \############M\$:           -X############
 \##########H;.      ,--.     =X##########
 :X######M;     -\$H\@M##MH%:    :H#######\@
   =%#M+=,   ,+\@#######M###H:    -=/M#%
   %M##\@+   .X##\$, ./+- ./###;    +M##%
   %####M.  /###=         \@##M.   X###%
   %####M.  ;M##H:.     =\$###X.   \$###%
   %####\@.   /####M\$-./\@#####:    %###%
   %H#M/,     /H###########\@:     ./M#%
  ;\$H##\@\@H:    .;\$HM#MMMH\$;,   ./H\@M##M\$=
 X#########%.      ..,,.     .;\@#########
 \###########H+:.           ./\@###########
 \##############/ ./%%%%+/.-M#############
 \##############H\$\@#######\@\@##############
 \##############X%########M\$M#############
 +M##########H:            .\$##########X=
EOC

```

### cower
```perl
##
## A cowering cow
##
$the_cow = <<EOC;
     $thoughts
      $thoughts
        ,__, |    | 
        ($eyes)\\|    |___
        (__)\\|    |   )\\_
         $tongue  |    |_w |  \\
             |    |  ||   *

             Cower....
EOC

```

### cowfee
```perl
$the_cow = <<EOC;
   $thoughts      {
    $thoughts  }   }   {
      {   {  }  }
       }   }{  {
      {  }{  }  }
     ( }{ }{  { )
    .-{   }   }-.
   ( ( } { } { } )
   |`-.._____..-'|
   |             ;--.
   |   (__)     (__  \\
   |   ($eyes)      | )  )
   |    \\/       |/  /
   |     $tongue      /  /
   |            (  /
   \\             y'
    `-.._____..-'
EOC

```

### cthulhu-mini
```perl
# Cthulhu
#
$the_cow = <<EOC;
  $thoughts
   $thoughts

      ^(;,;)^

EOC


```

### cube
```perl
# Cube
#
# from http://www.reddit.com/r/commandline/comments/2lb5ij/what_is_your_favorite_ascii_art/cltrase
#   also available at https://gist.github.com/th3m4ri0/6e3f631866da31d05030
# 
$the_cow = <<EOC;
   $thoughts
    $thoughts
       ____________
      /\\  ________ \\
     / /\\ \\______/\\ \\
    / / /\\ \\  / /\\ \\ \\
   / / /__\\ \\/ / /\\ \\ \\
  / /_/____\\ \\/_/__\\_\\ \\
  \\ \\ \\____/ / ________ \\
   \\ \\ \\  / / /\\ \\  / / /
    \\ \\ \\/ / /\\ \\ \\/ / /
     \\ \\/ / /__\\_\\/ / /
      \\  / /______\\/ /
       \\/___________/
EOC


```

### daemon
```perl
##
## 4.4 >> 5.4
##
$the_cow = <<EOC;
   $thoughts         ,        ,
    $thoughts       /(        )`
     $thoughts      \\ \\___   / |
            /- _  `-/  '
           (/\\/ \\ \\   /\\
           / /   | `    \\
           $eye $eye   ) /    |
           `-^--'`<     '
          (_.)  _  )   /
           `.___/`    /
             `-----' /
<----.     __ / __   \\
<----|====O)))==) \\) /====
<----'    `--' `.__,' \\
             |        |
              \\       /
        ______( (_  / \\______
      ,'  ,-----'   |        \\
      `--{__________)        \\/
EOC

```

### dalek-shooting
```perl
# Dalek
# from http://www.asciiworld.com/-Robots,24-.html (accessed 4/30/2014)
$the_cow = <<EOC;
                                    $thoughts
                                     $thoughts
                                                         ____                   
                                               [(=]|[==/   @  \\     
                                                      |--------|                
     *                                     *  .       ==========                
.  / *    .                         *   .* . * /.     ==========                
 / /  .                      *   .    *  \\. * /      ||||||||||||               
 =-=-=-=-=-=-----==-=--=-=--=-=-=-=---=--= -. %%%%%%[-- ||||||||||              
  \\  \\ .                             *  (===========[  /=========]              
.  \\   *  *                          .    /  * \\   |==============]             
         *                        *      *         C @ @ @ @ @ @ |D             
        *  *                          .           /              |              
                                         .       C  @ @ @  @ @  @ |D            
          *                          *          /                 |             
                                               C  @  @  @  @  @  @ |D           
                                              /                    |            
                                             C  @   @   @   @  @  @ |D          
                                            /                       |           
                                           |@@@@@@@@@@@@@@@@@@@@@@@@@|          
                                            -------------------------           
Modified from howard1\@vax.oxford.ac.uk
EOC

```

### dalek
```perl
# Dalek
# from http://www.ascii-art.de/ascii/def/dr_who.txt (accessed 4/30/2014)
$the_cow = <<EOC;
   $thoughts
    $thoughts
              ___
      D>=G==='   '.
            |======|
            |======|
        )--/]IIIIII]
           |_______|
           C O O O D
          C O  O  O D
         C  O  O  O  D
         C__O__O__O__D
snd     [_____________]
EOC

```

### default
```perl
$the_cow = <<"EOC";
        $thoughts   ^__^
         $thoughts  ($eyes)\\_______
            (__)\\       )\\/\\
             $tongue ||----w |
                ||     ||
EOC

```

### docker-whale
```perl
##
## docker whale
##
$the_cow = <<EOC;
         $thoughts
          $thoughts
                    ##        .
              ## ## ##       ==
           ## ## ## ##      ===
       /""""""""""""""""\___/ ===
  ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
       \______ o          __/
         \    \        __/
          \____\______/

EOC

```

### doge
```perl
##
## Doge
##
$the_cow = <<EOC;
   $thoughts
    $thoughts

           _                _
          / /.           _-//
         / ///         _-   /
        //_-//=========     /
      _///        //_ ||   ./
    _|                 -__-||
   |  __              - \\   \
  |  |#-       _-|_           |
  |            |#|||       _   |  
 |  _==_                       ||
- ==|.=.=|_ =                  |
|  |-|-  ___                  |
|    --__   _                /
||     ===                  |
 |                     _. //
  ||_         __-   _-  _|
     \_______/  ___/  _|
                   --*
EOC

```

### dolphin
```perl
# dolphin (tiny)
#
# from http://www.chris.com/ascii/index.php?art=animals/other%20(water)
$the_cow = <<EOC;
     $thoughts
      $thoughts
               ,
             __)\\_  
       (\_.-'    a`-.
  jgs  (/~~````(/~^^` 

EOC

```

### dragon-and-cow
```perl
##
## A dragon smiting a cow, possible credit to kube@csua.berkeley.edu
##
$the_cow = <<EOC;
                       $thoughts                    ^    /^
                        $thoughts                  / \\  // \\
                         $thoughts   |\\___/|      /   \\//  .\\
                          $thoughts  /O  O  \\__  /    //  | \\ \\           *----*
                            /     /  \\/_/    //   |  \\  \\          \\   |
                            \@___\@`    \\/_   //    |   \\   \\         \\/\\ \\
                           0/0/|       \\/_ //     |    \\    \\         \\  \\
                       0/0/0/0/|        \\///      |     \\     \\       |  |
                    0/0/0/0/0/_|_ /   (  //       |      \\     _\\     |  /
                 0/0/0/0/0/0/`/,_ _ _/  ) ; -.    |    _ _\\.-~       /   /
                             ,-}        _      *-.|.-~-.           .~    ~
            \\     \\__/        `/\\      /                 ~-. _ .-~      /
             \\____($eyes)           *.   }            {                   /
             (    (--)          .----~-.\\        \\-`                 .~
             //__\\\\$tongue\\__ Ack!   ///.----..<        \\             _ -~
            //    \\\\               ///-._ _ _ _ _ _ _{^ - - - - ~
EOC

```

### dragon
```perl
##
## The Whitespace Dragon
##
$the_cow = <<EOC;
      $thoughts                    / \\  //\\
       $thoughts    |\\___/|      /   \\//  \\\\
            /$eye  $eye  \\__  /    //  | \\ \\    
           /     /  \\/_/    //   |  \\  \\  
           \@_^_\@'/   \\/_   //    |   \\   \\ 
           //_^_/     \\/_ //     |    \\    \\
        ( //) |        \\///      |     \\     \\
      ( / /) _|_ /   )  //       |      \\     _\\
    ( // /) '/,_ _ _/  ( ; -.    |    _ _\\.-~        .-~~~^-.
  (( / / )) ,-{        _      `-.|.-~-.           .~         `.
 (( // / ))  '/\\      /                 ~-. _ .-~      .-~^-.  \\
 (( /// ))      `.   {            }                   /      \\  \\
  (( / ))     .----~-.\\        \\-'                 .~         \\  `. \\^-.
             ///.----..>        \\             _ -~             `.  ^-`  ^-_
               ///-._ _ _ _ _ _ _}^ - - - - ~                     ~-- ,.-~
                                                                  /.-~
EOC

```

### ebi_furai
```perl
#
# えびフライ
#

$the_cow = <<EOC;
  $thoughts
   $thoughts
      ,.,,､,..,､､.,､,､､.,_          ／i
    ;'`;、､:、..:、:,:,.::｀'::ﾞ":,'´ --i
    '､;:..: ,:.､.:',.:.::_.;..;:.‐'ﾞ

EOC

```

### elephant-in-snake
```perl
##
## Do we need to explain this?
##
$the_cow = <<EOC;
   $thoughts
    $thoughts              ....       
           ........    .      
          .            .      
         .             .      
.........              .......
..............................

Elephant inside ASCII snake
EOC

```

### elephant
```perl
##
## An elephant out and about
##
$the_cow = <<EOC;
 $thoughts     /\\  ___  /\\
  $thoughts   // \\/   \\/ \\\\
     ((    $eye $eye    ))
      \\\\ /     \\ //
       \\/  | |  \\/ 
        |  | |  |  
        |  | |  |  
        |   o   |  
        | |   | |  
        |m|   |m|  
EOC

```

### elephant2
```perl
# Elephant
$the_cow = <<EOC;
  $thoughts
   $thoughts                                 
      /  \\~~~/  \\         
     (    ..     )----,      
      \\__     __/      \\     
        )|  /)         |\\    
         | /\\  /___\\   / ^   
          "-|__|   |__|      
EOC

```

### explosion
```perl
# Explosion
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
    $thoughts
     $thoughts
            .+
             /M;
              H#@:              ;,
              -###H-          -@/
               %####\$.  -;  .%#X
                M#####+;#H :M#M.
..          .+/;%#########X###-
 -/%H%+;-,    +##############/
    .:\$M###MH\$%+############X  ,--=;-
        -/H#####################H+=.
           .+#################X.
         =%M####################H;.
            /@###############+;;/%%;,
         -%###################\$.
       ;H######################M=
    ,%#####MH\$%;+#####M###-/@####%
  :\$H%+;=-      -####X.,H#   -+M##@-
 .              ,###;    ;      =\$##+
                .#H,               :XH,
                 +                   .;-
EOC

```

### eyes
```perl
##
## Evil-looking eyes
##
$the_cow = <<EOC;
    $thoughts
     $thoughts
                                   .::!!!!!!!:.
  .!!!!!:.                        .:!!!!!!!!!!!!
  ~~~~!!!!!!.                 .:!!!!!!!!!UWWW\$\$\$ 
      :\$\$NWX!!:           .:!!!!!!XUWW\$\$\$\$\$\$\$\$\$P 
      \$\$\$\$\$##WX!:      .<!!!!UW\$\$\$\$"  \$\$\$\$\$\$\$\$# 
      \$\$\$\$\$  \$\$\$UX   :!!UW\$\$\$\$\$\$\$\$\$   4\$\$\$\$\$* 
      ^\$\$\$B  \$\$\$\$\\     \$\$\$\$\$\$\$\$\$\$\$\$   d\$\$R" 
        "*\$bd\$\$\$\$      '*\$\$\$\$\$\$\$\$\$\$\$o+#" 
             """"          """"""" 
EOC

```

### fat-banana
```perl
# fatter banana
# via https://www.reddit.com/r/cowsay/comments/3bkpwv/any_love_for_bananasay/
$the_cow = <<EOC;
           $thoughts
            $thoughts
        "-.. __      __.='>
         `.     """""   ,'
           "-..__   _.-"
   ~ ~~ ~ ~  ~   """  ~~  ~
EOC

```

### fat-cow
```perl
# fatter cow
# via https://www.reddit.com/r/cowsay/comments/39htd0/with_all_this_reddit_hype_what_about_a_little/
$the_cow = <<EOC;
  $thoughts
   $thoughts

    A__A
   ( OO )\\_----__
   (____)\\      )\\/\\
        ||      |
        ||`---w||
EOC

```

### fence
```perl
$the_cow = <<EOC;
                          $thoughts
                           $thoughts         __.----.___
           ||            ||  (\\(__)/)-'||      ;--` ||
          _||____________||___`($eyes)'___||______;____||_
          -||------------||----)  (----||-----------||-
          _||____________||___(o  o)___||______;____||_
          -||------------||----`--'----||-----------||-
           ||            ||     $tongue `|| ||| || ||     ||jgs
        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
EOC

```

### fire
```perl
# Fire
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
      $thoughts
       $thoughts
                     -\$-
                    .H##H,
                   +######+
                .+#########H.
              -\$############\@.
            =H###############\@  -X:
          .\$##################:  \@#\@-
     ,;  .M###################;  H###;
   ;\@#:  \@###################\@  ,#####:
 -M###.  M#################\@.  ;######H
 M####-  +###############\$   =\@#######X
 H####\$   -M###########+   :#########M,
  /####X-   =########%   :M########\@/.
    ,;%H\@X;   .\$###X   :##MM\@%+;:-
                 ..
  -/;:-,.              ,,-==+M########H
 -##################\@HX%%+%%\$%%%+:,,
    .-/H%%%+%%\$H\@###############M\@+=:/+:
/XHX%:#####MH%=    ,---:;;;;/%%XHM,:###\$
\$\@#MX %+;-                           .
EOC

```

### flaming-sheep
```perl
##
## The flaming sheep, contributed by Geordan Rosario (geordan@csua.berkeley.edu)
##
$the_cow = <<EOC;
  $thoughts            .    .     .   
   $thoughts      .  . .     `  ,     
    $thoughts    .; .  : .' :  :  : . 
     $thoughts   i..`: i` i.i.,i  i . 
      $thoughts   `,--.|i |i|ii|ii|i: 
           U${eyes}U\\.'\@\@\@\@\@\@`.||' 
           \\__/(\@\@\@\@\@\@\@\@\@\@)'  
             $tongue (\@\@\@\@\@\@\@\@)    
                `YY~~~~YY'    
                 ||    ||     
EOC

```

### fox
```perl
# Fox
# http://www.retrojunkie.com/asciiart/animals/foxes.htm
$the_cow = <<EOC;
$thoughts
 $thoughts
   /\\   /\\   Todd Vargo
  //\\\\_//\\\\     ____
  \\_     _/    /   /
   / * * \\    /^^^]
   \\_\\O/_/    [   ]
    /   \\_    [   /
    \\     \\_  /  /
     [ [ /  \\/ _/
    _[ [ \\  /_/
EOC

```

### ghost
```perl
# art by Joan G. Stark, https://en.wikipedia.org/wiki/Joan_Stark
$the_cow = <<"EOC";
     $thoughts     .-.
      $thoughts  .'   `.
       $thoughts :g g   :
        $thoughts: o    `.
        :         ``.
       :             `.
      :  :         .   `.
      :   :          ` . `.
       `.. :            `. ``;
          `:;             `:'
             :              `.
              `.              `.     .
                `'`'`'`---..,___`;.-'
EOC


```

### ghostbusters
```perl
##
## Ghostbusters!
##
$the_cow = <<EOC;
          $thoughts
           $thoughts
            $thoughts          __---__
                    _-       /--______
               __--( /     \\ )XXXXXXXXXXX\\v.
             .-XXX(   $eye   $eye  )XXXXXXXXXXXXXXX-
            /XXX(       U     )        XXXXXXX\\
          /XXXXX(              )--_  XXXXXXXXXXX\\
         /XXXXX/ (      O     )   XXXXXX   \\XXXXX\\
         XXXXX/   /            XXXXXX   \\__ \\XXXXX
         XXXXXX__/          XXXXXX         \\__---->
 ---___  XXX__/          XXXXXX      \\__         /
   \\-  --__/   ___/\\  XXXXXX            /  ___--/=
    \\-\\    ___/    XXXXXX              '--- XXXXXX
       \\-\\/XXX\\ XXXXXX                      /XXXXX
         \\XXXXXXXXX   \\                    /XXXXX/
          \\XXXXXX      >                 _/XXXXX/
            \\XXXXX--__/              __-- XXXX/
             -XXXXXXXX---------------  XXXXXX-
                \\XXXXXXXXXXXXXXXXXXXXXXXXXX/
                  ""VXXXXXXXXXXXXXXXXXXV""
EOC

```

### glados
```perl
# GLaDOS from Portal
# via http://pastebin.com/1AZwKrKp 
$the_cow = <<EOC;
  $thoughts
   $thoughts
       \#+ \@      \# \#              M#\@
 .    .X  X.%##\@;# \#   +\@#######X. \@#%
   ,==.   ,######M+  -#####%M####M-    \#
  :H##M%:=##+ .M##M,;#####/+#######% ,M#
 .M########=  =\@#\@.=#####M=M#######=  X#
 :\@\@MMM##M.  -##M.,#######M#######. =  M
             \@##..###:.    .H####. \@\@ X,
   \############: \###,/####;  /##= \@#. M
           ,M## ;##,\@#M;/M#M  \@# X#% X#
.%=   \######M## \##.M#:   ./#M ,M \#M ,#\$
\##/         \$## \#+;#: \#### ;#/ M M- \@# :
\#+ \#M\@MM###M-;M \#:\$#-##\$H# .#X \@ + \$#. \#
      \######/.: \#%=# M#:MM./#.-#  \@#: H#
+,.=   \@###: /\@ %#,\@  \##\@X \#,-#\@.##% .\@#
\#####+;/##/ \@##  \@#,+       /#M    . X,
   ;###M#\@ M###H .#M-     ,##M  ;\@\@; \###
   .M#M##H ;####X ,\@#######M/ -M###\$  -H
    .M###%  X####H  .\@\@MM\@;  ;\@#M\@
      H#M    /\@####/      ,++.  / ==-,
               ,=/:, .+X\@MMH\@#H  \#####\$=
EOC

```

### goat
```perl
##
## ejm97 http://www.ascii-art.de/ascii/ghi/goat.txt
##
$the_cow = <<EOC;
       $thoughts
        $thoughts
         $thoughts  _))
           > $eye\\     _~
           `;'\\\\__-' \\_
              | )  _ \\ \\
             / / ``   w w
            w w
EOC






```

### goat2
```perl
#
#	CodeGoat.io: https://github.com/danyshaanan/goatsay
#
$the_cow = <<EOC;
        $thoughts
         $thoughts
          )__(
         '|$eyes|'________/
          |__|         |
           $tongue||"""""""||
             ||       ||

EOC

```

### golden-eagle
```perl
# Golden Eagle (Marquette University mascot)
# 
$the_cow = <<EOC;
    $thoughts                                       ,:=+++++???++?+=+=
     $thoughts                               :+?????\$MUUUUUUUUUMO+??~
      $thoughts                         :+I??\$UUUUMUMMMMUUMUMUUMUUMM???I+:
       $thoughts                     ,+??+ZOUUMMUMMMUUUUUMUUUMUUUMUMUUMZI+?+:
        $thoughts                 ~I?+MMUMUUUMUUUOOUMMMMMMUUUUUMMMUUUUUUMMUM$??~
                       I?+7MMMMMUUO7?+?IUMMMMMMMMUUMUUMUUUUUUUUUUMMMUMO?I
                    ~I?+MMMUUUO????+?IOUZ7?,.......,+\$\$OUMUUUUUUMUMUUUMUU+I:
                   =??\$UMUUU7++??????II???????=.....,?OUUMMMUUMUUUUUUUMMUUU+=
                 +??UUMMM7??????+??+?????+??=,...\$MUMUUUUMUUMUUUUUUUUMMUM7II??=
               ,+?IUMMMI???III?++??+?????+~....... ......MUUMUUUUUUUUUMMU7?~
               IIUMMM+?+?IUUUUMUUM7I?????????????I?+=:......MUUMMUMUMMMMUUU+~
             :?+UMMU+?+?7UMMUUUZ7\$\$7????+++????????????=.....+UMUUUUMMMMMMUZ?
             ?+UMUM???+MMMMMU?++???????????++????????++????....OMMMUMMMMUMMUI:
            +\$MMUM?+?ZMMU:\$MM???OUUU+??+???+????????????????,...UMUMUUUUMMUMM?~
            IUUUU?I?OUUU,..UU?IMMUUMUUI???+?????????????????I,..:UUMUUUUMMUMU?+
            ?UUUMUM\$UMUU~..UUUUU\$,IUUUMM7+?????????????????+?I~..UUUUUUUMMMUU+?
            ?OUMUUUUMMUI+.?UUUU=...~UMMUU\$?????+???????????????..MUMUUUUUMUMU??
           :??IUUMMUMUMMOMUU7........OUUUMMU?I????????????????I..MUMUUUUMUUMU?+
         +IIUMUO.IUUUUUUO..............?UMUMUM7??????????????+?..UUMUUUUMUUMU?=
       ,IZMMU,.:UU7:..........,UUUMUZ....MUUMMO+???????????????..UUUUUUUUUUU?:
       IZUUU:..UUUI=....... IUUUUUMMUZ,.MMUMU$?+???????????????.MUUMUUUUMUMUI
     ,+IUUM..O=..........\$UUMMMUU?~....UMMUUI?????????????????=.UMMUUUUUMMU?+
     +?UMU~............OUMMUU~..... .UUMUMM+?????????????????=.UMMUMUMUUMOI=
     ?\$MU~...    ...:MUMU=~........,UUMMMUI+????????????????+IUMMMUUUUMUU?+
     +OMU....   ...?UMU=..:~~,.....MMMUUU+?+????????????????~MMUUUUMUMMU?+~   
     ?OMU~ .. ...?UMUUUMUMUMUMUMUUUMUUMUI???????????????+?+OUUMUUMMMMUIUUUMO,
     ??UMU~.....\$MUUUOM???UMMUUUMMMUUMM7?++????????????++OMMMUMUUMMUI??UMIU+~
     :?7UUU\$...UMMM?I~,  +?MMUUMMMMMMUU?????????????+??\$UMMMMUUUMU\$?: ,??I?:
       ?IMUMUUZMU+?,      =?UMMMMMMMMO??????????????+UMUMMUMUMUU?I~
        ?+\$MUMUMU??        ?MMMMMMMMU??+???????????IUMMMUUUMUUZ?=
          ,+???ZUO?~       +ZUMMUMUMU???++??+???IUMMUMUMMUUO??~
               ,,:~=       ,?UMUMUMU???+??+?+?7UMUMUMUMUI??:
                            ?UMUMMMM?+??++?ZUUMUMUMUZ++?,                  
                            ?UMMMMMO+???MMUMUMUMUMOII=,                       
                            ?UUUMUUZOMUUMUMMUMM+??=
                            ?UMMUMMUUUMUMM\$???~
                           ,?UMUUMUUU\$?+?~:                                  
                           :IUUUM?+?I=:                                  
                           ????~,
EOC

```

### hand
```perl
##
## これが私の本当の姿だ！
##  
##
$the_cow = <<EOC;
       $thoughts
        $thoughts
                           __ 
                  l^ヽ    /  }    _
                  |  |   /  /   ／  )
                  |  |  /  /  ／  ／ _
                  j. し'  / ／  ／ ／  )
                 /  .＿__ ´  ／ ／  ／
                /   {  /:｀ヽ ｀¨ ／
               /     ∨::::::ﾊ   ／
              |廴     ＼:::ノ}  /
    {￣￣￣￣ヽ  廴     ｀ー'  ー-､
    ヽ ＿＿_   ＼ 廴        ＿＿＿ﾉ
        ／       ＼ 辷_´￣
      ／           ﾍ￣
    ／             ,ﾍ
                  /、ﾍ
                 /＼__ﾉ
EOC


```

### happy-whale
```perl
# happy whale
#
# modified from http://www.chris.com/ascii/index.php?art=animals/other%20(water)
$the_cow = <<EOC;
   $thoughts
    $thoughts
     $thoughts
        __ \ / __
       /  \\ | /  \\
           \\|/
       _.---v---.,_
      /            \\  /\\__/\\
     /              \\ \\_  _/
     |__ @           |_/ /
      _/                / 
      \\       \\__,     /  
   ~~~~\\~~~~~~~~~~~~~~`~~~

EOC

```

### hedgehog
```perl
##
## A cute little hedgehog
##
$the_cow = <<EOC;
  $thoughts
   $thoughts ..:::::::::.
    ::::::::::::::
   /. `::::::::::::
  O__,_:::::::::::'
EOC

```

### hellokitty
```perl
##
## Hello Kitty
##
$the_cow = <<EOC;
  $thoughts
   $thoughts
      /\\_)o<
     |      \\
     | $eye . $eye|
      \\_____/
         $tongue
EOC

```

### hippie
```perl
$the_cow = <<EOC;
                       $thoughts              ___
                        $thoughts            ///\\\\\\/----
                         $thoughts           ||//\\///\\\\\\\\
                          $thoughts         /`-.__\\\\\\\\///|
                           $thoughts       /_  _   `--._|
                               ___-`---.___     |
                          ----------       `-.__|
                       ----------( \\.-.$eye $eye;_  \\\\\\\\\\\\
                      ------------| `-'-.(_)--/\\\\\\\\\\
                     /////------//|   `-'       )\\\\\\\\\\\\
                     /////------///\\  `--'\\  /"\\\\\\\\\\\\\\\\\\\\
                     ////--------///\\  `-' /\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\   .-.  _
                      //////------////>---'\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\  | | / )
        _              ////////////// |__| )\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\  | |/ /
       / `.       _    ////////.-'  >\\    <-._.--.\\\\\\\\\\\\\\\\\\  _|__ /_
      (    \\  . .' )    /////// ( .- (    )() ( )_)\\\\\\\\\\\\\\  / __)-' )
       `-   | |/          //// ( ) ( )|--'() ( ) \\\\\\\\\\\\\\\\   \\  `(.-')
          .---/ _()        /// ( ) () |  /() ( )  \\\\\\\\\\\\     > ._>-'
        ()+8 8 |            |  ( )( ) | /( ) ( )   \\        / \\/
        ()+8 8/-()__        /  ( )( ) \\/ ( ) ( )\\   \\      /\\ /
          |8 8|     `.     |   () ( ).--.( ) ( )-\\   \\    /   |
        ()+||||-() (_/    _/   /| ()/ || \\ )  ()()\\   \\__/   /
        .-`||||          /\\\\  / / ()|/  \\ ()     \\ `.  /|   |
       (_  ||||        .'   _/-/  ()\\/||\\/()     \\-. \\     /
           ||( \\_    .'    ( )/  ( ) `--' ( )     > ) `.  /
        .--|_|\\_ \\ .'    .'( )_  ( )-.___.-( )  (  )    ""
        `.__)-.( /.'\\  .'   (  )'_)-.______( ).-')'
          (___)|  \\ .-'      `--'`-._.---._.(_))-'
          (__)|| +-)'           |   /_.--.\\    |
          (__)||-'              `._|`-'  ) )  _|
            |||||                |  `.`-'.'--' /
            |||||               .'    | |   .\\|
            |||||             .'   _.-|_|     \\
            |||||            /   .'.-'  \\\\     |
            ||||||         .'     /      \\     \\
             |||||        /     .'        \\     \\
             |||||      .'     /           |    |
            _|||||----./     .'            \\     \\
         .-' |||||   `/     /               \\    |
       .'     |||||   (    /                |    |
      /       |||||   |    |\\                \\   |
      |     .'|||||.  |    ||                |    )
       \\    | |||||\\  |    |/                |    \\
        \\   | ||||||  |    |                 /    |
        |    `.||||' /     |                |     \\
        |      ||||  |     \\                |      |
        /      ||||| |      |\\             /       |
       /       |||||_/      | \\            |        \\
      /      ------'|       |  |           |        |
     |      |___.---|        \\ |           /        |
     |             /         | |          |         \\
     |             |         \\/           |          |
     |             /          |           |          |
      \\           |           |           |          |
       `.        /             \\          |           \\
         `--.___`-_            |_         |           |
           .-.__.-''-,_         -         |           \\_'
          <`.         '.-//|-/``        (_)          _.-'
           `._-.____.-'.|   /            '//, ,\\.-'`` |--.
              `-.____.' |__/               '''\\      -'/ |
                                               `.   _.// |
                                                 `-.__.-'

VK
EOC

```

### hiya
```perl
$the_cow = <<EOC;
           $thoughts     (      )
            $thoughts    ~(^^^^)~
             $thoughts    ) $eyes \\~_          |\\
              $thoughts  /     | \\        \\~ /
                ( 0  0  ) \\        | |
                 ---___/~  \\       | |
                  /'__/ |   ~-_____/ |
   o          _   ~----~      ___---~
     O       //     |         |
            ((~\\  _|         -|
      o  O //-_ \\/ |        ~  |
           ^   \\_ /         ~  |
                  |          ~ |
                  |     /     ~ |
                  |     (       |
                   \\     \\      /\\
                  / -_____-\\   \\ ~~-*
                  |  /       \\  \\       .==.
                  / /         / /       |  |
                /~  |      //~  |       |__|         W<
                ~~~~        ~~~~
EOC

```

### hiyoko
```perl
##
## ひよ子
##
$the_cow = <<EOC;
  $thoughts
   $thoughts
    $thoughts
      ,､ ,._
      ﾉ ・  ヽ
     / :::   i  
    / :::    ﾞ､
   ,i:::       `ｰ-､
   |:::           i
   !::::..        ﾉ
    `ー――――'" 
EOC

```

### homer
```perl
# Homer Simpson
#
# from http://www.reddit.com/r/textfiles/comments/2s9ybk/random_ascii_art/
#
$the_cow = <<EOC;
            $thoughts
             $thoughts                __ 
                   _ ,___,-'",-=-. 
       __,-- _ _,-'_)_  (""`'-._\\ `. 
    _,'  __ |,' ,-' __)  ,-     /. | 
  ,'_,--'   |     -'  _)/         `\\ 
,','      ,'       ,-'_,`           : 
,'     ,-'       ,(,-(              : 
     ,'       ,-' ,    _            ; 
    /        ,-._/`---'            / 
   /        (____)(----. )       ,' 
  /         (      `.__,     /\\ /, 
 :           ;-.___         /__\\/| 
 |         ,'      `--.      -,\\ | 
 :        /            \\    .__/ 
  \\      (__            \\    |_ 
   \\       ,`-, *       /   _|,\\ 
    \\    ,'   `-.     ,'_,-'    \\ 
   (_\\,-'    ,'\\")--,'-'       __\\ 
    \\       /  // ,'|      ,--'  `-. 
     `-.    `-/ \\'  |   _,'         `. 
        `-._ /      `--'/             \\ 
-hrr-      ,'           |              \\ 
          /             |               \\ 
       ,-'              |               / 
      /                 |             -'
EOC

```

### hypno
```perl
$the_cow =<<"EOC"
  $thoughts
     ___        _--_
    /    -    /     \\
   ( $eyes   \\  (    $eyes )
   |  $eyes _;\\-/|  $eyes _|
    \\___/######\\___/\\
      /##############\\
     /  ######   ##  #|
    /  ##@##@##       |
   /    ######     ##  \\
 <______-------___\\  . //_
    |       ____  | | //# \\__~__
     \\      $tongue    \\  //###  \\   \\
      |             /\'  ##  ##  ##\\   __--~--_
       \\_________- /\\ )    ^     ##|--########\\
  /--~-_\\________/_  |          #@##|#######Y##|
 | \\ `  /|       /O/ ( ###  \')    ##/######/###/
 \\  \\  | |       --  |  ###        /LLLLL--###/
  \\_ \\/  |            \\_   \\    ) /####_____--
 ___ /    \\           /     |   _-####\\
(___/     -\\_________/     / -- |#####@@@@@@\'_
 (__\\_      __,) (.___     ,/    /#####      `@@
      | -\\\\-          //-//      @@  @@@@@.
      | | \\\\_       _// //      @\'       \'@@.
      (.)   \\_)    / / //                   @@@
                  (_) (_\'
EOC

```

### ibm
```perl
#
# International Business Machines
#

$the_cow = << EOC;
  $thoughts
   $thoughts

■■■■■   ■■■■■■■■     ■■■■■       ■■■■■
■■■■■   ■■■■■■■■■■   ■■■■■■     ■■■■■■
 ■■■     ■■■   ■■■    ■■■■■■   ■■■■■■
 ■■■     ■■■■■■■■     ■■■■■■■ ■■■■■■■
 ■■■     ■■■■■■■■     ■■■ ■■■■■■■ ■■■
 ■■■     ■■■   ■■■    ■■■  ■■■■■  ■■■
■■■■■   ■■■■■■■■■■   ■■■■   ■■■   ■■■■
■■■■■   ■■■■■■■■     ■■■■    ■    ■■■■
EOC

```

### iwashi
```perl
##
## いわし
##  
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
         ＿＿＿＿ ＿＿＿＿＿__
      ｨ''  ＠ :. ,! ，， ， ，￣￣ ¨` ‐-            ＿＿
       ＼    ノ   i            ’ ’’ ’’､_;:`:‐.-_-‐ニ＝=彳
         ｀ ＜. _  .ｰ ､                       !三  ＜
                 ｀¨  ‐= . ＿＿＿_.. ﾆ=-‐‐`'´｀ﾐ､   三＞
                                                 ￣￣
EOC

```

### jellyfish
```perl
# jellyfish
#
# from http://ascii.co.uk/art/jellyfish
$the_cow = <<EOC;
     $thoughts
      $thoughts

         .-;\':\':\'-.
        {\'.\'.\'.\'.\'.}
         )        \'`.
        \'-. ._ ,_.-=\'
          `). ( `);(
          (\'. .)(,\'.)
           ) ( ,\').(
          ( .\').\'(\').
          .) (\' ).(\'
           '  ) (  ).
            .\'( .)\'
              .).\'
jgs

EOC

```

### karl_marx
```perl
##
## Karl Marx
##  
##
$the_cow = <<EOC;
       $thoughts
        $thoughts
                   ,―ヾヽヽ/ｖへ／⌒ー
                , ⌒ヽ ヽ ヽ / ／ ノ  ⌒ヽ、
              / ／ヾ,ゞ -ゞゞゞ､_ ⌒  ノ ヽ
            ／  ／            `ヾ  ー   ミヽ
          ,/   /                   ヾ ＼  ヽﾐ
         /    /                      ゞ      ヽ
         i   /                       /      ＼
        /    -=ﾆヽ､,_  ,,,,;r;==-     ヾ  ヾミ ヽ
        | ;: `ゞﾂヽ〉^`ヾだ'=-､_        i    彡 ヽ
        i ,   /::::/     `'''"""        ﾉ  ゞ ヾ ヽ
        } ;  |    人､,;-,'^            /    くヾ  ）
        /    彡ノノノﾉﾉﾉ(((((        ／ﾍミ        /
       /     /ﾉﾉﾉﾉﾉ,.-―ミヽヾヾヾヾヾヾ     _ノ`ｰ'"
      ,i          -ー‐ `ゞ           ヽ   ヽ
      彡彡                        ミ       ヽ
''""￣彡      /   /   /   /            ミ   ﾂ＼
      ＜    /   /   /   /        ヾ   ヾ  ノﾉﾉ
        '―彡                         ｒー'"
            ヾノ人,,.r--､ノノノノノり'"
EOC

```

### kilroy
```perl
# Kilroy
# from http://www.ascii-art.de/ascii/jkl/kilroy.txt (accessed 8/14/2014)
$the_cow = <<EOC;
     $thoughts 
      $thoughts
           ,,,
          (0 0)
   +---ooO-(_)-Ooo---+
   |                 |
EOC


```

### king
```perl
# King (Chess piece)
#
# from http://www.chessvariants.org/d.pieces/ascii.html
#   by David Moeser
#
$the_cow = <<EOC;
 $thoughts
  $thoughts
    .::.
    _::_
  _/____\\_
  \\      /
   \\____/
   (____)
    |  |
    |__|
   /    \\
  (______)
 (________)
 /________\\
EOC

```

### kiss
```perl
##
## A lovers' empbrace
##
$the_cow = <<EOC;
     $thoughts
      $thoughts
             ,;;;;;;;,
            ;;;;;;;;;;;,
           ;;;;;'_____;'
           ;;;(/))))|((\\
           _;;((((((|))))
          / |_\\\\\\\\\\\\\\\\\\\\\\\\
     .--~(  \\ ~))))))))))))
    /     \\  `\\-(((((((((((\\\\
    |    | `\\   ) |\\       /|)
     |    |  `. _/  \\_____/ |
      |    , `\\~            /
       |    \\  \\           /
      | `.   `\\|          /
      |   ~-   `\\        /
       \\____~._/~ -_,   (\\
        |-----|\\   \\    ';;
       |      | :;;;'     \\
      |  /    |            |
      |       |            |
EOC

```

### kitten
```perl
# Kitten
#
# based on rfksay by Andrew Northern
# http://robotfindskitten.org/aw.cgi?main=software.rfk#rfksay
#
$the_cow = <<EOC;
   $thoughts
    $thoughts

     |\\_/|
     |o o|__
     --*--__\\
     C_C_(___)
EOC

```

### kitty
```perl
##
## A kitten of sorts, I think
##
$the_cow = <<EOC;
     $thoughts
      $thoughts
       ("`-'  '-/") .___..--' ' "`-._
         ` $eye_ $eye  )    `-.   (      ) .`-.__. `)
         (_Y_.) ' ._   )   `._` ;  `` -. .-'
      _.. `--'_..-_/   /--' _ .' ,4
   ( i l ),-''  ( l i),'  ( ( ! .-'    
EOC

```

### knight
```perl
# Knight (Chess piece)
#
# from http://www.chessvariants.org/d.pieces/ascii.html
#   by David Moeser
#
$the_cow = <<EOC;
 $thoughts
  $thoughts
  __/"""\\
 ]___ 0  }
     /   }
   /~    }
   \\____/
   /____\\
  (______)
EOC

```

### koala
```perl
##
## From the canonical koala collection
##
$the_cow = <<EOC;
  $thoughts
   $thoughts
       ___  
     {~$eye_$eye~}
      ( Y )
     ()~*~()   
     (_)-(_)   
EOC

```

### kosh
```perl
##
## It's a Kosh Cow!
##
$the_cow = <<EOC;
    $thoughts
     $thoughts
      $thoughts
  ___       _____     ___
 /   \\     /    /|   /   \\
|     |   /    / |  |     |
|     |  /____/  |  |     |     
|     |  |    |  |  |     |
|     |  | {} | /   |     |
|     |  |____|/    |     |
|     |    |==|     |     |
|      \\___________/      |
|                         |
|                         |
EOC

```

### lamb
```perl
$the_cow = <<EOC;
                 $thoughts
                  $thoughts  _,._
                 __.'   _)
                <_,)'.-"$eye\\
                  /' (    \\
      _.-----..,-'   (`"--^
     //              |   $tongue
    (|   `;      ,   |  
      \\   ;.----/  ,/ 
       ) // /   | |\\ \\
       \\ \\\\`\\   | |/ /
        \\ \\\\ \\  | |\\/
EOC

```

### lamb2
```perl
$the_cow = <<EOC;
 $thoughts
  $thoughts
  ,-''''-.
 (.  ,.   L        ___...__
 /$eye} ,-`  `'-==''``        ''._
//{                           '`.
\\_,X ,                         : )
 $tongue 7                          ;`
    :                  ,       /
     \\_,                \\     ;
       Y   L_    __..--':`.    L
       |  /| ````       ;  y  J
       [ j J            / / L ;
       | |Y \\          /_J  | |
       L_J/_)         /_)   L_J
      /_)               sk /_)
EOC

```

### lightbulb
```perl
# lightbulb
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
$thoughts
 $thoughts
         ,=;%\$%%\$X%%%%;/%%%%;=,
     ,/\$\$+:-                -:+\$\$/,
   :X\$=                          =\$X:
 ;M%.                              .%M;
+#/                                  /#+
\##                                    M#
H#,                     =;+/;,       ,#X
.HM-       :\@X+%H:   .%M%- .M#.     -M\@.
  /#%.     \@#-  ,H\@--MH, .;\@\$-    .%#+
   .\$M;    .+\@X;, MM#\@:/\$X;.     ;M\$,
     =\@H,     ,:+%H#M%;-       ,H\@=
      .\$#;        -#H         =#\$
        %#;        \#M        ;#%
         H#-       \##       -#H
         ;#+       \##       +#;
          ;H+;;;;;;HH;;;;;;+H/
           =H#\@HHHHHHHHHH\@#H=
           =\@#H%%%%%%%\$HH\@#\@=
           =\@#X%%%%%%%\$M###\@=
               =+%XHHX%+=
EOC

```

### lobster
```perl
# Lobster
#   lobster jgs   10/96
#   http://ascii.co.uk/art/lobster
$the_cow = <<EOC;
             $thoughts
              $thoughts
                             ,.---._
                   ,,,,     /       `,
                    \\\\\\\\   /    '\\_  ;
                     |||| /\\/``-.__\\;'
                     ::::/\\/_
     {{`-.__.-'(`(^^(^^^(^ 9 `.========='
    {{{{{{ { ( ( (  (   (-----:=
     {{.-'~~'-.(,(,,(,,,(__6_.'=========.
                     ::::\\/\\
                     |||| \\/\\  ,-'/,
                    ////   \\ `` _/ ;
                   ''''     \\  `  .'
                             `---'
EOC

```

### lollerskates
```perl
# LOLLERSKATES
$the_cow = <<EOC;
   $thoughts
    $thoughts
        /\\O
         /\\/
        /\\
       /  \\
      LOL LOL
:-D LOLLERSKATES :-D
EOC

```

### luke-koala
```perl
##
## From the canonical koala collection
##
$the_cow = <<EOC;
  $thoughts
   $thoughts          .
       ___   //
     {~$eye_$eye~}// 
      ( Y )K/  
     ()~*~()   
     (_)-(_)   
     Luke    
     Skywalker
     koala   
EOC

```

### mailchimp
```perl
# MailChimp
#
# view-source:http://mailchimp.com/
$the_cow = <<EOC;
$thoughts
 $thoughts
    ______
   / ___M ]__
C{ ( o o )}
    {     ••
      \\___
      ----´
EOC

```

### maze-runner
```perl
# maze-runner.cow
#
#   a guy running through an ASCII maze
#   found at http://pip.readthedocs.org/en/user_builds/pip/rtd-builds/latest/installing/
#
$the_cow = <<EOC;
    $thoughts
     $thoughts
      $thoughts
       \\
        \\
         \\
    \\     \\                     /
     \\     \\                   /
      \\     \\                 /
       ]     \\               [    ,'|
       ]      \\              [   /  |
       ]___               ___[ ,'   |
       ]  ]\\             /[  [ |:   |
       ]  ] \\           / [  [ |:   |
       ]  ]  ]         [  [  [ |:   |
       ]  ]  ]__     __[  [  [ |:   |
       ]  ]  ] ]\\ _ /[ [  [  [ |:   |
       ]  ]  ] ] (#) [ [  [  [ :===='
       ]  ]  ]_].nHn.[_[  [  [
       ]  ]  ]  HHHHH. [  [  [
       ]  ] /   `HH("N  \\ [  [
       ]__]/     HHH  "  \\[__[
       ]         NNN         [
       ]         N/"         [
       ]         N H         [
      /          N            \\
     /           q,            \\
    /                           \\
EOC


```

### mech-and-cow
```perl
$the_cow = <<EOC;
      $thoughts                            |     |
       $thoughts                        ,--|     |-.
                         __,----|  |     | |
                       ,;::     |  `_____' |
                       `._______|    i^i   |
                                `----| |---'| .
                           ,-------._| |== ||//
                           |       |_|P`.  /'/
                           `-------' 'Y Y/'/'
                                     .==\ /_\
   ^__^                             /   /'|  `i
   ($eyes)\_______                   /'   /  |   |
   (__)\       )\/\             /'    /   |   `i
    $tongue ||----w |           ___,;`----'.___L_,-'`\__
       ||     ||          i_____;----\.____i""\____\
EOC

```

### meow
```perl
##
## A meowing tiger?
##
$the_cow = <<EOC;
  $thoughts
   $thoughts ,   _ ___.--'''`--''//-,-_--_.
      \\`"' ` || \\\\ \\ \\\\/ / // / ,-\\\\`,_
     /'`  \\ \\ || Y  | \\|/ / // / - |__ `-,
    /\$eye"\\  ` \\ `\\ |  | ||/ // | \\/  \\  `-._`-,_.,
   /  _.-. `.-\\,___/\\ _/|_/_\\_\\/|_/ |     `-._._)
   `-'``/  /  |  // \\__/\\__  /  \\__/ \\
    $tongue  `-'  /-\\/  | -|   \\__ \\   |-' |
          __/\\ / _/ \\/ __,-'   ) ,' _|'
         (((__/(((_.' ((___..-'((__,'
EOC

```

### milk
```perl
##
## Milk from Milk and Cheese
##
$the_cow = <<EOC;
 $thoughts     ____________ 
  $thoughts    |__________|
      /           /\\
     /           /  \\
    /___________/___/|
    |          |     |
    |  ==\\ /== |     |
    |   $eye   $eye  | \\ \\ |
    |     <    |  \\ \\|
   /|          |   \\ \\
  / |  \\_____/ |   / /
 / /|    $tongue    |  / /|
/||\\|          | /||\\/
    -------------|   
        | |    | | 
       <__/    \\__>
EOC

```

### minotaur
```perl
$the_cow = <<"EOC";
        $thoughts   ^__^
         $thoughts  ($eyes)
            (__)
           /-||-\\
           \\|\\/|/
            o==o 
            ||||
            ()()
EOC

```

### mona-lisa
```perl
# Mona Lisa
#
# from http://www.heartnsoul.com/ascii_art/mona_lisa_ascii.htm
$the_cow = <<EOC;
          $thoughts
           $thoughts

!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!>''''''<!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!'''''`             ``'!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!''`          .....         `'!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!'`      .      :::::'            `'!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!'     .   '     .::::'                `!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!'      :          `````                   `!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!        .,cchcccccc,,.                       `!!!!!!!!!!!!
!!!!!!!!!!!!!!!     .-"?\$\$\$\$\$\$\$\$\$\$\$\$\$\$c,                      `!!!!!!!!!!!
!!!!!!!!!!!!!!    ,ccc\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$,                     `!!!!!!!!!!
!!!!!!!!!!!!!    z\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$;.                    `!!!!!!!!!
!!!!!!!!!!!!    <\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$:.                    `!!!!!!!!
!!!!!!!!!!!     \$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$h;:.                   !!!!!!!!
!!!!!!!!!!'     \$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$h;.                   !!!!!!!
!!!!!!!!!'     <\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$                   !!!!!!!
!!!!!!!!'      `\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$F                   `!!!!!!
!!!!!!!!        c\$\$\$\$???\$\$\$\$\$\$\$P""  """??????"                      !!!!!!
!!!!!!!         `"" .,.. "\$\$\$\$F    .,zcr                            !!!!!!
!!!!!!!         .  dL    .?\$\$\$   .,cc,      .,z\$h.                  !!!!!!
!!!!!!!!        <. \$\$c= <\$d\$\$\$   <\$\$\$\$=-=+"\$\$\$\$\$\$\$                  !!!!!!
!!!!!!!         d\$\$\$hcccd\$\$\$\$\$   d\$\$\$hcccd\$\$\$\$\$\$\$F                  `!!!!!
!!!!!!         ,\$\$\$\$\$\$\$\$\$\$\$\$\$\$h d\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$                   `!!!!!
!!!!!          `\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$<\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$'                    !!!!!
!!!!!          `\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$"\$\$\$\$\$\$\$\$\$\$\$\$\$P>                     !!!!!
!!!!!           ?\$\$\$\$\$\$\$\$\$\$\$\$??\$c`\$\$\$\$\$\$\$\$\$\$\$?>'                     `!!!!
!!!!!           `?\$\$\$\$\$\$I7?""    ,\$\$\$\$\$\$\$\$\$?>>'                       !!!!
!!!!!.           <<?\$\$\$\$\$\$c.    ,d\$\$?\$\$\$\$\$F>>''                       `!!!
!!!!!!            <i?\$P"??\$\$r--"?""  ,\$\$\$\$h;>''                       `!!!
!!!!!!             \$\$\$hccccccccc= cc\$\$\$\$\$\$\$>>'                         !!!
!!!!!              `?\$\$\$\$\$\$F""""  `"\$\$\$\$\$>>>''                         `!!
!!!!!                "?\$\$\$\$\$cccccc\$\$\$\$??>>>>'                           !!
!!!!>                  "\$\$\$\$\$\$\$\$\$\$\$\$\$F>>>>''                            `!
!!!!!                    "\$\$\$\$\$\$\$\$???>'''                                !
!!!!!>                     `"""""                                        `
!!!!!!;                       .                                          `
!!!!!!!                       ?h.
!!!!!!!!                       \$\$c,
!!!!!!!!>                      ?\$\$\$h.              .,c
!!!!!!!!!                       \$\$\$\$\$\$\$\$\$hc,.,,cc\$\$\$\$\$
!!!!!!!!!                  .,zcc\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$
!!!!!!!!!               .z\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$
!!!!!!!!!             ,d\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$          .
!!!!!!!!!           ,d\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$         !!
!!!!!!!!!         ,d\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$        ,!'
!!!!!!!!>        c\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$.       !'
!!!!!!''       ,d\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$>       '
!!!''         z\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$>
!'           ,\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$>             ..
            z\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$'           ;!!!!''`
            \$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$F       ,;;!'`'  .''
           <\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$>    ,;'`'  ,;
           `\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$F   -'   ,;!!'
            "?\$\$\$\$\$\$\$\$\$\$?\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$F     .<!!!'''       <!
         !>    ""??\$\$\$?C3\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$""     ;!'''          !!!
       ;!!!!;,      `"''""????\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$\$""   ,;-''               ',!
      ;!!!!<!!!; .                `"""""""""""    `'                  ' '
      !!!! ;!!! ;!!!!>;,;, ..                  ' .                   '  '
     !!' ,;!!! ;'`!!!!!!!!;!!!!!;  .        >' .''                 ;
    !!' ;!!'!';! !! !!!!!!!!!!!!!  '         -'
   <!!  !! `!;! `!' !!!!!!!!!!<!       .
   `!  ;!  ;!!! <' <!!!! `!!! <       /
  `;   !>  <!! ;'  !!!!'  !!';!     ;'
   !   !   !!! !   `!!!  ;!! !      '  '
  ;   `!  `!! ,'    !'   ;!'
      '   /`! !    <     !! <      '
           / ;!        >;! ;>
             !'       ; !! '
          ' ;!        > ! '

EOC

```

### moofasa
```perl
##
## MOOfasa.
##
$the_cow = <<EOC;
       $thoughts    ____
        $thoughts  /    \\
          | ^__^ |
          | ($eyes) |______
          | (__) |      )\\/\\
           \\____/|----w |
                ||     ||

	         Moofasa
EOC

```

### mooghidjirah
```perl
$the_cow = <<EOC;
 $thoughts       $thoughts      $thoughts      
  $thoughts        ^__^  $thoughts        
    ^__^   ($eyes)   ^__^  
    ($eyes)   (__)   ($eyes)   
    (__)    $tongue    (__)   
oyo/:$tongue            $tongue:/oy+
/mmmmm+   syyyyo  `ommmmm/
 smmmmms. -ymmy. .smmmmmo 
 `+dmmmmd+``::``+dmmmmd+  
   -ymmmmmh/``+hmmmmmy-   
    `/hmmmmmhhmmmmmh/`    
      `/hmmmmmmmmh/`      
        `/hmmmmmd/        
      `oo.`/dmmmmdo`      
     `ymmd+``ommmmmy`     
     smmmmd-  /mmmmms     
    -mmmmm+    ommmmm-    
    -ooooo`    .ooooo.     
EOC

```

### moojira
```perl
$the_cow = <<EOC;
     $thoughts              
      $thoughts    /ss/           
   `oys:  .dmmd`  :syo`   
   /dmmy   .//.   hmmd:   
    -/:`          `:/-    
oyo/:.     ^__^     .:/oy+
/mmmmm+   <($eyes\)>  `ommmmm/
 smmmmms. -(__). .smmmmmo 
 `+dmmmmd+``$tongue``+dmmmmd+  
   -ymmmmmh/``+hmmmmmy-   
    `/hmmmmmhhmmmmmh/`    
      `/hmmmmmmmmh/`      
        `/hmmmmmd/        
      `oo.`/dmmmmdo`      
     `ymmd+`VVmmmmmy`     
     smmmmd-  /mmmmms     
    -mmmmm+    ommmmm-    
    -ooooo`    .ooooo.    
EOC

```

### moose
```perl
$the_cow = <<EOC;
  $thoughts
   $thoughts   \\_\\_    _/_/
    $thoughts      \\__/
           ($eyes)\\_______
           (__)\\       )\\/\\
            $tongue ||----- |
               ||     ||
EOC

```

### mule
```perl
# Mule
#
# based on mule from http://rossmason.blogspot.com/2008/10/friday-ascii-art.html 
#
$the_cow = <<EOC;
     $thoughts
      $thoughts 
  /\\          /\\                               
 ( \\\\        // )                              
  \\ \\\\      // /                               
   \\_\\\\||||//_/                                
     / _  _ \\/                                 
                                               
     |(o)(o)|\\/                                
     |      | \\/                               
     \\      /  \\/_____________________         
      |____|     \\\\                  \\\\        
     /      \\     ||                  \\\\       
     \\ 0  0 /     |/                  |\\\\      
      \\____/ \\    V           (       / \\\\     
       / \\    \\     )          \\     /   \\\\    
      / | \\    \\_|  |___________\\   /     "" 
                  ||  |     \\   /\\  \\          
                  ||  /      \\  \\ \\  \\         
                  || |        | |  | |         
                  || |        | |  | |         
                  ||_|        |_|  |_|         
                 //_/        /_/  /_/          
EOC

```

### mutilated
```perl
##
## A mutilated cow, from aspolito@csua.berkeley.edu
##
$the_cow = <<EOC;
       $thoughts   \\_______
 v__v   $thoughts  \\   O   )
 ($eyes)      ||----w |
 (__)      ||     ||  \\/\\
  $tongue
EOC

```

### nyan
```perl
# Nyan Cat
#
# from http://www.reddit.com/r/commandline/comments/2lb5ij/what_is_your_favorite_ascii_art/clt4ybl
#
$the_cow = <<EOC;
       $thoughts
        $thoughts

+      o     +              o   
    +             o     +       +
o          +
    o  +           +        +
+        o     o       +        o
-_-_-_-_-_-_-_,------,      o 
_-_-_-_-_-_-_-|   /\\_/\\  
-_-_-_-_-_-_-~|__( ^ .^)  +     +  
_-_-_-_-_-_-_-''  ''      
+      o         o   +       o
    +         +
o        o         o      o     +
    o           +
+      +     o        o      +    
EOC


```

### octopus
```perl
# octopus
#   http://www.ascii-art.de/ascii/mno/octopus.txt
$the_cow = <<EOC;
        $thoughts               ___
         $thoughts           .-'   `'.
                    /         \\
                    |         ;
                    |         |           ___.--,
           _.._     |0) ~ (0) |    _.---'`__.-( (_.
    __.--'`_.. '.__.\\    '--. \\_.-' ,.--'`     `""`
   ( ,.--'`   ',__ /./;   ;, '.__.'`    __
   _`) )  .---.__.' / |   |\\   \\__..--""  """--.,_
  `---' .'.''-._.-'`_./  /\\ '.  \\ _.-~~~````~~~-._`-.__.'
        | |  .' _.-' |  |  \\  \\  '.               `~---`
         \\ \\/ .'     \\  \\   '. '-._)
          \\/ /        \\  \\    `=.__`~-.
     jgs  / /\\         `) )    / / `"".`\\
    , _.-'.'\\ \\        / /    ( (     / /
     `--~`   ) )    .-'.'      '.'.  | (
            (/`    ( (`          ) )  '-;
             `      '-;         (-'
EOC

```

### okazu
```perl
#
# おかず
#

$the_cow = <<EOC;
  $thoughts
   $thoughts                _, _ ,､
    $thoughts          , - ´      `--、
             ノ               丶
           ／                  `､_
         ,´                        、
        ,'                          丶
       ﾉ                             ヽ
    ＿;＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿',＿
    ヽ三三三三三三三三三三三三三三三三三ﾉ
      ヽ                              /
       ヽ三三三三三三三三三三三三三三/
         ＼                        ／
           ＼三三三三三三三三三三／
             `＜              ＞´
               ｀丁三三三三丁´
     ＿          ｀ ｰ----‐ ´
  ／::/＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿＿_
（;;;ﾌ ｰ─----＝＝ === ニニニ 二二二三三三｣

         ＿|＿ ＼  ＿ｌ＿＼  _＿|＿_ヽヽ
          _|＿       ｜ヽ     __|
        ／ |  ヽ     ﾉ  │   (__|
        ＼ノ  ノ    ﾉ ヽﾉ     _ノ 
EOC

```

### owl
```perl
##
## An owl
##
$the_cow = <<EOC;
         $thoughts
          $thoughts
           ___
          (o o)
         (  V  )
        /--m-m-
EOC

```

### pawn
```perl
# Pawn (Chess piece)
#
# from http://www.chessvariants.org/d.pieces/ascii.html
#   by David Moeser
#
$the_cow = <<EOC;
 $thoughts
  $thoughts
     __
    (  )
     ||
    /__\\
   (____)
EOC

```

### periodic-table
```perl
$the_cow = <<EOC;
$thoughts
 $thoughts
   1A   2A                                         3A  4A  5A  6A  7A  8A
  -----                                                               -----
1 | H |                                                               |He |
  |---+----                                       --------------------+---|
2 |Li |Be |                                       | B | C | N | O | F |Ne |
  |---+---|                                       |---+---+---+---+---+---|
3 |Na |Mg |3B  4B  5B  6B  7B |    8B     |1B  2B |Al |Si | P | S |Cl |Ar |
  |---+---+---------------------------------------+---+---+---+---+---+---|
4 | K |Ca |Sc |Ti | V |Cr |Mn |Fe |Co |Ni |Cu |Zn |Ga |Ge |As |Se |Br |Kr |
  |---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---|
5 |Rb |Sr | Y |Zr |Nb |Mo |Tc |Ru |Rh |Pd |Ag |Cd |In |Sn |Sb |Te | I |Xe |
  |---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---|
6 |Cs |Ba |Lu |Hf |Ta | W |Re |Os |Ir |Pt |Au |Hg |Tl |Pb |Bi |Po |At |Rn |
  |---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---+---|
7 |Fr |Ra |Lr |Rf |Db |Sg |Bh |Hs |Mt |Ds |Rg |Cn |Nh |Fl |Mc |Lv |Ts |Og |
  -------------------------------------------------------------------------
              -------------------------------------------------------------
   Lanthanide |La |Ce |Pr |Nd |Pm |Sm |Eu |Gd |Tb |Dy |Ho |Er |Tm |Yb |Lu |
              |---+---+---+---+---+---+---+---+---+---+---+---+---+---+---|
   Actinide   |Ac |Th |Pa | U |Np |Pu |Am |Cm |Bk |Cf |Es |Fm |Md |No |Lr |
              -------------------------------------------------------------
EOC

```

### personality-sphere
```perl
# Personality Sphere from Portal/Portal 2
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
   $thoughts
    $thoughts
      .-+\$H###MM\@MMMMM##\@\$+-,. ....
-\@\$+%\$+%HX+--..  .  . .,:X\$/+/++\$#:
-#MXH\$=                      \$HXH#:
 .--,:#+   ,+\$HMX =\@\@X%, . .X#:,,,
     =#\@\$H :####H =####;,M%\$#X
     X###\$ \$####X =####H %###X
    ;###X /###\@\$: ,+HM##H.+###;
   :###;,X##%=;%H\@H\$;-;M#\@-;###/
  ,M##;.\@##;-H#######M=.M##-:###-
  ;##M ;##X \@###H-=\@###.;##X H##;
  ;##M./##X.\@###H:/M###-=##X X##;
  -###;,M##:,\@########+-H##; \@##-
   %##M==\@##%==%HMH%::/M##+.X##+
    %###/./###X+: -+\$M##M=,X##+
     X###X X####H +#####% \@##H
     :###H %####H +#####; X##;
     /#\$.  -HM##H /###\@+.  +#\$. .
/HX%\$X:      .,-, .-,.      =XX\$H\@-
/#H+/+%+/+;=.          .=/%;;/;;+#+
 ..  .,-:XM#MM\@\@\@\@\@\@H\@\@M#\@+=,.   ,,
EOC

```

### pinball-machine
```perl
# Pinball machine
#
# from http://ascii.co.uk/art/pinball
$the_cow = <<EOC;
    $thoughts
     $thoughts
              /\\
             <  \\
             |\\  \\
             | \\  \\
             | .\\  >
             |  .\\/|
             |   .||
             |    ||
            / \\   ||
           /,-.\\: ||
          /,,  `\\ ||
         /,  ', `\\||
        /, *   ''/ |
       /,    *,'/  |
      /,     , /   |
     / :    , /   .|
    /\\ :   , /   /||
   |\\ \\ .., /   / ||
   |.\\ \\ . /   /  ||
   |  \\ \\ /   /   ||
   |   \\ /   /    |'
   |\\o '|o  /
   ||\\o |  /
   || \\ | /
   ||  \\|/
   |'   ||
        ||
        ||
        |'
EOC

```

### psychiatrichelp
```perl
$the_cow = <<EOC;
        $thoughts         ____________________
         $thoughts       |                    |
          $thoughts      |     PSYCHIATRIC    |
           $thoughts     |        HELP        |
            $thoughts    |____________________|
             $thoughts   ||  ,-..'``.        ||
              $thoughts  || (,-..'`. )       ||
                 ||   )-c - `)\\      ||
   ,.,._.-.,_,.,-||,.(`.--  ,`',.-,_,||.-.,.,-,._.
              ___||____,`,'--._______||
             |`._||______`'__________||
             |   ||     __           ||
             |   ||    |.-' ,|-      ||
   _,_,,..-,_|   ||    ._)) `|-      ||,.,_,_.-.,_
            . `._||__________________||   ____    .
     .              .           .     . <.____`>
   .SSt  .      .     .      .    .   _.()`'()`'  .
EOC
   
```

### psychiatrichelp2
```perl
use utf8;
$the_cow = <<EOC;
 $thoughts      .------------------------.
  $thoughts     |       PSYCHIATRIC      |
   $thoughts    |         HELP  5¢       |
    $thoughts   |________________________|
     $thoughts  ||     .-"""--.         ||
      $thoughts ||    /        \\.-.     ||
        ||   |     ._,     \\    ||
        ||   \\_/`-'   '-.,_/    ||
        ||   (_   (' _)') \\     ||
        ||   /|           |\\    ||
        ||  | \\     __   / |    ||
        ||   \\_).,_____,/}/     ||
      __||____;_--'___'/ (      ||
     |\\ ||   (__,\\\\    \\_/------||
     ||\\||______________________||
     ||||                        |
     ||||       THE DOCTOR       |
     \\|||         IS [IN]   _____|
      \\||                  (______)
 jgs   `|___________________//||\\\\
                           //=||=\\\\
                           `  ``  `
EOC

```

### pterodactyl
```perl
# pterodactyl.cow
#
#   a pterodactyl with its mouth open
#
$the_cow = <<EOC;
    $thoughts
     $thoughts
      $thoughts
                                                                                 -/- 
                                                                              -/ --/    
                                                                            /- -  /     
                                                                         //      /      
                                                                        /       /       
                                                                      //       /        
                                                                    //        /         
                                                                  //          /         
                                                                ///           /         
                                                               //            /          
                                                              //            /           
                                                             //          . ./           
                                                             //       .    /            
                                                             //    .      /             
                                                             //  .       /              
                                                            // .         /              
                                                          (=>            /              
                                                         (==>            /              
                                                          (=>            /              
             -_                                           //.           /               
             \\\\-_                                        //   .         /               
              \\ \\_-_                                     //     .       /               
               \\_ \\_--_                                 //        . . . /               
                 \\_ \\_ -_                              //              /                
                   \\_ \\_ (O)-___                      //               /                
                     \\ _\\   __  --__                  /                /                
                     _/    \\  ----__--____          //                 /                
                   _/  _/   \\       -------       //                  /                 
                 _/ __/ \\\\   \\\\                  /                   /                  
               _/ _/      \\\\   \\\\              //                   /                   
              -__/          \\\\   \\\\\\          //                   /                    
                              \\\\    \\\\\\\\\\\\\\\\\\//   -                /                    
                                \\\\         _/         -            /                    
                                  \\\\                      -        \\                    
                                    \\\\\\                       -     \\                   
                                        \\\\                       -   \\                  
                                          \\\\\\                         \\--__             
                                           | \\\\                            \\__________  
                                            |  \\\\\\\\                ___      _________-\\\\
                                            |    \\\\\\\\\\                \\--__/____        
                                            |        \\\\\\\\________---\\-    ______-----   
                                             |                   /    \\--  \\_______     
                                             |                   /       \\-_________\\   
                                             \\                   /                  \\\\  
                                             \\                 ./                       
                                             \\            .     /                       
                                              \\        .       /                        
                                              \\    .           //                       
                                              \\                /                        
                                              |__              /                        
                                              \\==              /                        
                                               \\\\              \\                        
                                                \\\\  .          \\                        
                                                  \\\\    .  .   \\                        
                                                   \\           .\\                       
                                                   \\\\            \\                      
                                                     \\           \\                      
                                                      \\\\          \\                     
                                                        \\\\         \\                    
                                                          \\         \\--                 
                                                           \\\\          \\                
                                                             \\\\         \\\\\\\\            
                                                               \\\\\\\\_________\\\\\\         
EOC


```

### queen
```perl
# Queen (Chess piece)
#
# from http://www.chessvariants.org/d.pieces/ascii.html
#   by David Moeser
#
$the_cow = <<EOC;
 $thoughts
  $thoughts
     ()
   <~~~~>
    \\__/
   (____)
    |  |
    |  |
    |__|
   /____\\
  (______)
 (________)
EOC

```

### R2-D2
```perl
# R2-D2
#
# from http://www.ascii-art.de/ascii/s/starwars.txt
$the_cow = <<EOC;
   $thoughts
    $thoughts
         _____
       .\'/L|__`.
      / =[_]O|` \\
      |\"+_____\":|
    __:='|____`-:__
   ||[] ||====| []||
   ||[] | |=| | []||
   |:||_|=|U| |_||:|
   |:|||]_=_ =[_||:| LS
   | |||] [_][]C|| |
   | ||-\'\"\"\"\"\"`-|| |
   /|\\\\_\\_|_|_/_//|\\
  |___|   /|\\   |___|
  `---\'  |___|  `---\'
         `---'
EOC

```

### radio
```perl
# radio from Portal
# via http://pastebin.com/1AZwKrKp
$the_cow = <<EOC;
     $thoughts
      $thoughts
                    ;=
                    /=
                    ;=
                    /=
                    ;=
                    /=
                    ;=
                    /=
             ,--==-:\$;
         ,/\$@#######\@X+-
      ./@###############X=
     /M#####X+/;;;;+H#####\$.
    %####M/;+H\@XX@@%;;\@####\@,
   +####H=+##\$,--,=M#X-%####\@.
  -####X,X\@HHXH##MXHXXH-+####\$
  X###\@.X/\$M\$:####\$=\@X/X,X####-
 .####:+\$:##\@:####\$:##H/X=####%
 -%%\$%,+==%\$+-\$+:\$;-\$\$%-+,/\$%%+
 -/+%%X\$XX\$\$\$\$\$\$\$%\$\$\$%\$X\$X\$%+/-
EOC

```

### ren
```perl
##
## Ren 
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
    ____  
   /# /_\\_
  |  |/$eye\\$eye\\
  |  \\\\_/_/
 / |_   |  
|  ||\\_ ~| 
|  ||| \\/  
|  |||_    
 \\//  |    
  ||  |    
  ||_  \\   
  \\_|  o|  
  /\\___/   
 /  ||||__ 
    (___)_)
EOC

```

### renge
```perl
##
## Nyanpasu~
##
$the_cow = <<EOC;
     $thoughts               _
      $thoughts            ´   ＼   __
       $thoughts        ／ ／⌒\\ | ／   ＼
   f|{r、       | /     '|/ ／⌒＼＼
   ||J |        \\/＞--＜\\/ /--    |
(＼|`` し]ﾄ----／          ⌒` ＼| /
 ＼      ﾉ\\   /                ＼|/\\   --、___
  ゛    /  ＼/      /     |         \\/_       ﾉ
   \\、/\\_／/ｲ    ,/'|    /\\ 、        Ⅵ   __／
    [\\/   \\/_|   /\\|/|   |-]  、     く-く
    |      \\/|  |/___ﾉ\\  /\\___ \\     /   ＼
    {/      <|小| _ﾒﾘ  \\/  _ﾒﾘ` \\   ｜|   |
     \\        ｜| \\/ｿ      \\/ｿ  ﾉ / /\\|＼_/
      \\       ｜|              /_ｲ\\/
       \\      ｜|     /ヽ      / /ﾉ
        \\     ｜/\\   └-     ,/ /'
         \\    ｜ |／>> r -=≦{{/ /ﾆ=_
          \\   人 | ／ｨ|     /ﾚ/__   ﾉﾆ-、
           ＼   \\|/  Xﾉ    / /   入//⌒Yﾊ
             \\  /し ｜`---' //  /  \\ﾆﾆﾆﾉ|
              ＼/  / \\  --ｱ ｜  |   | _]|
               ｜ /   \\/\\/  ｜  |   |___|
               r勺    ｜_｜ ｜  |   |  ||
               |`7    ｜ ｜ ｜  |   |   |
EOC

```

### robot
```perl
# Robot
#
# based on rfksay by Andrew Northern
# http://robotfindskitten.org/aw.cgi?main=software.rfk#rfksay
#
$the_cow = <<EOC;
  $thoughts
   $thoughts

     [-]
     (+)=C
     | |
     OOO
EOC

```

### robotfindskitten
```perl
# Robot finds kitten <3
#
# based on rfksay by Andrew Northern
# http://robotfindskitten.org/aw.cgi?main=software.rfk#rfksay
#
$the_cow = <<EOC;
  $thoughts
   $thoughts

    [-]   |\\_/|
    (+)=C |o o|__
    | |   --*--__\\
    OOO   C_C_(___)
EOC

```

### roflcopter
```perl
$the_cow = <<EOC;
   $thoughts
    $thoughts
 ROFL:ROFL:ROFL:ROFL
         _^___
 L    __/   $eyes \\    
LOL===__        \\ 
 L      \\________]
         I   I  $tongue
        --------/
EOC

```

### rook
```perl
# Rook (Chess piece)
#
# from http://www.chessvariants.org/d.pieces/ascii.html
#   by David Moeser
#
$the_cow = <<EOC;
 $thoughts
  $thoughts

   WWWWWW
    |  |
    |  |
    |__|
   /____\\
  (______)
EOC

```

### sachiko
```perl
#
# プロデューサーさんは独特の変わったセンスをしてますね！
#
$the_cow = <<EOC;
     $thoughts
      $thoughts
       $thoughts
             , -――- 、
          ／          ヽ、
        /爻ﾉﾘﾉﾊﾉﾘlﾉ ゝ  l
     ＜ﾉﾘﾉ‐'    ｰ  ﾘ ＞ }
        l ﾉ ┃    ┃ l ﾉ  ﾉ
        l人   r‐┐   !ﾉ＾)
           ゝ ` ´ ‐＜´
EOC

```

### satanic
```perl
##
## Satanic cow, source unknown.
##
$the_cow = <<EOC;
     $thoughts
      $thoughts  (__)  
         (\\/)  
  /-------\\/    
 / | 666 ||$tongue  
*  ||----||      
   ~~    ~~      
EOC

```

### seahorse-big
```perl
# large seahorse
#
# adapted from http://www.chris.com/ascii/index.php?art=animals/other%20(water)
$the_cow = <<EOC;
     $thoughts
      $thoughts
                  ,
         ___     /^\\   ,
        `\  \'...`   \\_/^\\
          ) ~     ',    /__,
         /       ,.    ,, /___,
        (  .-.   \'.\'. /// ___/
         ) .-.\'  .`.`///-.\'.
        / ( o )  .\"\". ====) \\
       (   \'-`   \\  |\'~~~`  u\\,
        \\ _~  .\"\"\"` |~|^u^ u^(\"\"
        //  ."     /~/^ u^ u^\
       // ."      /~  u^ u  ^u\      _
      // ."      /~/U^ U^ U^ ^(     / )
     /` ."       |~  U^ U^ ^ U^\   /) _)
   ./` ."        |~|^ U^ ^U ^ U(  / _  _)
  ;.`."          |~ ^U ^ U^ U ^/ /)_ =  _)
   \"\"            |~|^ ^U ^ ^ U(_/_    )- _)
                 |~ U ^ ^U ^U ^ )   =    _)
                 \\~|^ U U^ U ^ =  ~ )  - _)
                  \\ U ^U ^ ^U^_)     =  _)
                   \",^U^ ^U ^/ \\)_~   -_)
                     \".u^u ^|   \\_  = _)
                      ).u ^u|    \\)  _)
                      \\u ^u^(     \\__)
                       )^u ^u\\
                       \\u ^u ^|
             ____       )^u ^u|
          ,-`    '-.    )u ^u^|
         /  .---. ' \\  / ^ u^/
        |  ;  `  '  | /u^u ^/
        |  ;  '-` . `:u^u^u/
        \\.\'^\'._   _.`u ^.-`
         \\_.~=_```-.^.-\"
           \'\"------\"`

EOC

```

### seahorse
```perl
# seahorse
#
# adapted from http://www.chris.com/ascii/index.php?art=animals/other%20(water)
$the_cow = <<EOC;
   $thoughts
    $thoughts

      (\\(\\/
  .-._)oo  '_
  \'---.     .\'\\
       )    \\.-\'\\
      /__ ;     (
      |__ : /'._/
       \\_  (
       .,)  )
       \'-.-\'

EOC

```

### sheep
```perl
##
## The non-flaming sheep.
##
$the_cow = <<EOC
  $thoughts
   $thoughts
       __     
      U${eyes}U\\.'\@\@\@\@\@\@`.
      \\__/(\@\@\@\@\@\@\@\@\@\@)
        $tongue (\@\@\@\@\@\@\@\@)
           `YY~~~~YY'
            ||    ||
EOC

```

### shikato
```perl
$the_cow = <<EOC;
  $thoughts
   $thoughts

     Lｰ'{r ｧjｰノ
      _`)-ﾑ{
    /´::( ･)ヽ-- ､
   {::::::::::::::}
   ゝ:::::.ノー-
     しｿ¨UU
EOC

```

### shrug
```perl
$the_cow = <<EOC;
  $thoughts
¯\\_(ツ)_/¯
EOC

```

### skeleton
```perl
##
## This 'Scowleton' brought to you by one of 
## {appel,kube,rowe}@csua.berkeley.edu
##
$the_cow = <<EOC;
          $thoughts      (__)      
           $thoughts     /$eyes|  
            $thoughts   (_"_)*+++++++++*
                   //I#\\\\\\\\\\\\\\\\I\\
                   I[I|I|||||I I `
                   I`I'///'' I I
                   I I       I I
                   ~ ~       ~ ~
                     Scowleton
EOC

```

### small
```perl
##
## A small cow, artist unknown
##
$eyes = ".." unless ($eyes);
$the_cow = <<EOC;
       $thoughts   ,__,
        $thoughts  ($eyes)____
           (__)    )\\
            $tongue||--|| *
EOC

```

### smiling-octopus
```perl
# 
$the_cow = <<EOC;
      $thoughts
       $thoughts
        $thoughts                                     ,
                                            ,o
                                            :o
                   _....._                  `:o
                 .\'       ``-.                \\o
                /  _      _   \\                \\o
               :  /*\\    /*\\   )                ;o
               |  \\_/    \\_/   /                ;o
               (       U      /                 ;o
                \\  (\\_____/) /                  /o
                 \\   \\_m_/  (                  /o
                  \\         (                ,o:
                  )          \\,           .o;o\'           ,o\'o\'o.
                ./          /\\o;o,,,,,;o;o;\'\'         _,-o,-\'\'\'-o:o.
 .             ./o./)        \\    \'o\'o\'o\'\'         _,-\'o,o\'         o
 o           ./o./ /       .o \\.              __,-o o,o\'
 \\o.       ,/o /  /o/)     | o o\'-..____,,-o\'o o_o-\'
 `o:o...-o,o-\' ,o,/ |     \\   \'o.o_o_o_o,o--\'\'
 .,  ``o-o\'  ,.oo/   \'o /\\.o`.
 `o`o-....o\'o,-\'   /o /   \\o \\.                       ,o..         o
   ``o-o.o--      /o /      \\o.o--..          ,,,o-o\'o.--o:o:o,,..:o
                 (oo(          `--o.o`o---o\'o\'o,o,-\'\'\'        o\'o\'o
                  \\ o\\              ``-o-o\'\'\'\'
   ,-o;o           \\o \\
  /o/               )o )  Carl Pilcher
 (o(               /o /                |
  \\o\.       ...-o\'o /              \\   |
    \\o`o`-o\'o o,o,--\'       ~~~~~~~~\\~~|~~~~~~~~~~~~~~~~~~~~~~~~~~~~
      ```o--\'\'\'                       \\| /
                                       |/
 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~|~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                                       |
 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

EOC

```

### snoopy
```perl
##
## acsii picture From: kwok@menpachi.nmfs.hawaii.edu (William Kwok)
## from http://www.ascii-art.de/ascii/s/snoopy.txt
$the_cow = <<EOC;
 $thoughts
  $thoughts          , ----.
   $thoughts        -  -     `
      ,__.,'           \\
    .'                 *`
   /       $eye   $eye     / **\\
  .                 / ****.
  |    mm           | ****|
   \\                | ****|
    ` ._______      \\ ****/
              \\      /`---'
               \\___(
               /~~~~\\
              /      \\
             /      | \\
            |       |  \\
  , ~~ .    |, ~~ . |  |\\
 ( |||| )   ( |||| )(,,,)`
( |||||| )-( |||||| )    | ^
( |||||| ) ( |||||| )    |'/
( |||||| )-( |||||| )___,'-
 ( |||| )   ( |||| )
  ` ~~ '     ` ~~ '
EOC

```

### snoopyhouse
```perl
##
## acsii picture from http://www.ascii-art.de/ascii/s/snoopy.txt
##
$the_cow = <<EOC;
       $thoughts
        $thoughts       __---__                         ______
         $thoughts     /    ___\\_             o  O  O _(      )__
              /====(_____\\___---_  o        _(           )_
             |                    \\        (_  AI-YA!!!!   )
             |                     |@        (_  Shot      _)
              \\       ___         /           (__  Again!__)
 \\ __----____--_\\____(____\\_____/                (______)
==|__----____--______|
 /              /    \\____/)_
              /        ______)
             /           |  |
            |           _|  |
       ______\\______________|______
      /                    *   *   \\
     /_____________*____*___________\\
     /   *     *                    \\
    /________________________________\\
    / *                              \\
   /__________________________________\\
        |                        |
        |________________________|
        |                        |
        |________________________|
EOC

```

### snoopysleep
```perl
##
## picture from http://www.ascii-art.de/ascii/ab/beagle.txt
## 
$the_cow = <<EOC;
 $thoughts
  $thoughts
   $thoughts     O_      __)(
       ,'  `.   (_".`.
      :      :    /|`
      |      |   ((|_  ,-.
      ; -   /:  ,'  `:(( -\\
     /    -'  `: ____ \\\\\\-:
    _\\__   ____|___  \\____|_
   ;    | |        '-`      :
  :_____|:|__________________:
  ;     |:|                  :
 :      |:|                   :
 ;_______`'___________________:
:                              :
|______________________________|
 `---.--------------------.---'
     |____________________|
     |                    |
     |____________________|
     |                    |
   _\\|_\\|_\\/(__\\__)\\__\\//_|(_
EOC

```

### spidercow
```perl
$the_cow = <<EOC;
          $thoughts     (
           $thoughts     )
            $thoughts   (
         /\\  .-""""-.  /\\
        //\\\\/  ,,,,  \\//\\\\
        |/\\| ,;;;;;;, |/\\|
        //\\\\\\;-""""-;///\\\\
       //  \\/   ..   \\/  \\\\
      (| ,-_| \\ || / |_-, |)
        //`__(\\(__)/)__`\\\\
       // /.-\\`($eyes)'/-.\\ \\\\
      (\\ |)   ')  ('   (| /)
       ` (|   (o  o)   |) `
         \\)    `--'    (/
                $tongue
EOC

```

### squid
```perl
#
# これスプラトゥーン感あるね
#

$the_cow = <<EOC;
    $thoughts
     $thoughts                                                           ＿＿＿ノ^l
      $thoughts                                            ＿,,ノ``ｰ-'￣￣        ｌ
                                                 く                       /
                                                  `ヽ,   __､-'           /
                                                    __＞‐´               |
                                           ._,;‐''``              ,     /
                                         _;"                     /     /
                                       ／                       /     く
                                     ／                        /       |
                                   ／                        ／       ｌ
                                 ノ                        ／￣ヽ     /
                                /                        ／     ） _ノ
                            ,r'″ヽ、                   ／        ￣
                           /      ヽ                 ／
                        ＿ﾉ        `r            _､‐'
                      ／          _l,_       _､‐'
                 __,r'          ／r;;,ヽ   ／
               ,/              ｜.;●,;;|  ノ
              ノ ／  ／／       ヽ､!!!ﾞﾉ "
            ／ ／／／  ／／___,r''"￣
           / ／ / / /／ / /
      ___／／/／／／ ／／/
  ／￣＿_／／／/ / ／／／
 l ／´___／／／／／／ /
 しレ"／／/ /  ／／//／
      / ,/ / ／／／ /
      ﾚ'   ﾚ'／／ ／
           ／l｜l/
          ｜|ﾚ'lノ
           レ'
EOC

```

### squirrel
```perl
$the_cow = <<EOC;
  $thoughts
     $thoughts
                  _ _
       | \__/|  .~    ~.
       /$eyes `./      .'
      {o__,   \    {
        / .  . )    \
        `-` '-' \    }
       .(   _(   )_.'
      '---.~_ _ _|
                                                     
EOC

```

### stegosaurus
```perl
##
## A stegosaur with a top hat?
##
$the_cow = <<EOC;
$thoughts                             .       .
 $thoughts                           / `.   .' " 
  $thoughts                  .---.  <    > <    >  .---.
   $thoughts                 |    \\  \\ - ~ ~ - /  /    |
         _____          ..-~             ~-..-~
        |     |   \\~~~\\.'                    `./~~~/
       ---------   \\__/                        \\__/
      .'  $eye    \\     /               /       \\  " 
     (_____,    `._.'               |         }  \\/~~~/
      `----.          /       }     |        /    \\__/
            `-.      |       /      |       /      `. ,~~|
                ~-.__|      /_ - ~ ^|      /- _      `..-‘ / \\  /\\
                     |     /        |     /     ~-.     `-/ _ \\/__\\
                     |_____|        |_____|         ~ - . _ _ _ _ _>
EOC

```

### stimpy
```perl
##
## Stimpy!
##
$the_cow = <<EOC;
  $thoughts     .    _  .    
   $thoughts    |\\_|/__/|    
       / / \\/ \\  \\  
      /__|$eye||$eye|__ \\ 
     |/_ \\_/\\_/ _\\ |  
     | | (____) | ||  
     \\/\\___/\\__/  // 
     (_/         ||
      |          ||
      |          ||\\   
       \\        //_/  
        \\______//
       __ || __||
      (____(____)
EOC

```

### sudowoodo
```perl
# Sudowoodo (Pokémon)
#
# https://gist.github.com/rzabcik/9233650/
#
$the_cow = <<EOC;
     $thoughts
      $thoughts
     _              __
    / `\\  (~._    ./  )
    \\__/ __`-_\\__/ ./
   _ \\ \\/  \\   \\ |_   __
 (   )  \\__/ -^    \\ /  \\
  \\_/ "  \\  | o  o  |.. /  __
       \\. --' ====  /  || /  \\
         \\   .  .  |---__.\\__/
         /  :     /   |   |
         /   :   /     \\_/
      --/ ::    (
     (  |     (  (____
   .--  .. ----**.____)
   \\___/
EOC

```

### supermilker
```perl
##
## A cow being milked, probably from Lars Smith (lars@csua.berkeley.edu)
##
$the_cow = <<EOC;
  $thoughts   ^__^
   $thoughts  ($eyes)\\_______        ________
      (__)\\       )\\/\\    |Super |
       $tongue ||----W |       |Milker|
          ||    UDDDDDDDDD|______|
EOC

```

### surgery
```perl
##
## A cow operation, artist unknown
##
$the_cow = <<EOC;
          $thoughts           \\  / 
           $thoughts           \\/  
               (__)    /\\         
               ($eyes)   O  O        
               _\\/_   //         
         *    (    ) //       
          \\  (\\\\    //       
           \\(  \\\\    )                              
            (   \\\\   )   /\\                          
  ___[\\______/^^^^^^^\\__/) o-)__                     
 |\\__[=======______//________)__\\                    
 \\|_______________//____________|                    
     |||      || //||     |||
     |||      || @.||     |||                        
      ||      \\/  .\\/      ||                        
                 . .                                 
                '.'.`                                

            COW-OPERATION                           
EOC

```

### tableflip
```perl
$the_cow = <<EOC;
  $thoughts
(╯°□°）╯︵ ┻━┻
EOC

```

### taxi
```perl
# Taxi cab
#
# from http://ascii.co.uk/art/taxi
$the_cow = <<EOC;
     $thoughts
      $thoughts
                   [\\
              .----' `-----.
             //^^^^;;^^^^^^`\\
     _______//_____||_____()_\\________
    /826    :      : ___              `\\
   |>   ____;      ;  |/\\><|   ____   _<)
  {____/    \\_________________/    \\____}
       \\ '' /                 \\ '' /
 jgs    '--'                   '--'
EOC

```

### telebears
```perl
##
## A cow performing an unnatural act, artist unknown.
##
$the_cow = <<EOC;
      $thoughts                _
       $thoughts              (_)   <-- TeleBEARS
        $thoughts   ^__^       / \\
         $thoughts  ($eyes)\\_____/_\\ \\
            (__)\\  you  ) /
             $tongue ||----w ((
                ||     ||>> 
EOC

```

### template
```perl
# 
$the_cow = <<EOC;
$thoughts
 $thoughts
EOC

```

### threader
```perl
$the_cow = <<EOC;
       $thoughts
        $thoughts
         $thoughts
             ＿＿＿＿
           ／＿＿＿＿＼
         ／／ (⌒ ⌒ ヽ＼＼
        ｜｜  ﾉz(⌒ )| ｜｜
        ｜｜ <   (_ノ ｜｜
        ｜｜  L_／ )  ｜｜
         ＼＼ /＿／  ／／
           ＼⌒ )  (⌒ ／
           ／／    ＼＼
           ＼＼_  _／／
             ﾍ＿)(＿/
             ｜＝＝｜
              ＼三／
                ∧
              ／  ＼
              ＼  ／
                Ｖ
EOC

```

### threecubes
```perl
# Three cubes
#
# from http://www.reddit.com/r/commandline/comments/2lb5ij/what_is_your_favorite_ascii_art/cltcqs1
#   also available at https://gist.github.com/th3m4ri0/6e3f631866da31d05030
# 
$the_cow = <<EOC;
  $thoughts
   $thoughts
        ____________
       /\\  ________ \\
      /  \\ \\______/\\ \\
     / /\\ \\ \\  / /\\ \\ \\
    / / /\\ \\ \\/ / /\\ \\ \\
   / / /__\\ \\ \\/_/__\\_\\ \\__________
  / /_/____\\ \\__________  ________ \\
  \\ \\ \\____/ / ________/\\ \\______/\\ \\
   \\ \\ \\  / / /\\ \\  / /\\ \\ \\  / /\\ \\ \\
    \\ \\ \\/ / /\\ \\ \\/ / /\\ \\ \\/ / /\\ \\ \\
     \\ \\/ / /__\\_\\/ / /__\\ \\ \\/_/__\\_\\ \\
      \\  /_/______\\/_/____\\ \\___________\\
      /  \\ \\______/\\ \\____/ / ________  /
     / /\\ \\ \\  / /\\ \\ \\  / / /\\ \\  / / /
    / / /\\ \\ \\/ / /\\ \\ \\/ / /\\ \\ \\/ / /
   / / /__\\ \\ \\/_/__\\_\\/ / /__\\_\\/ / /
  / /_/____\\ \\_________\\/ /______\\/ /
  \\ \\ \\____/ / ________  __________/
   \\ \\ \\  / / /\\ \\  / / /
    \\ \\ \\/ / /\\ \\ \\/ / /
     \\ \\/ / /__\\_\\/ / /
      \\  / /______\\/ /
       \\/___________/
EOC


```

### toaster
```perl
# Toaster
#   http://ascii.co.uk/art/toaster 
$the_cow = <<EOC;
   $thoughts                     .___________.
    $thoughts                    |           |
     $thoughts    ___________.   |  |    /~\\ |
         / __   __  /|   | _ _   |_| |
        / /:/  /:/ / |   !________|__!
       / /:/  /:/ /  |            |
      / /:/  /:/ /   |____________!
     / /:/  /:/ /    |
    / /:/  /:/ /     |
   /  ~~   ~~ /      |
   |~~~~~~~~~~|      |
   |    ::    |     /
   |    ==    |    /
   |    ::    |   /
   |    ::    |  /
   |    ::  @ | /
   !__________!/
EOC

```

### tortoise
```perl
# Tortoise
# from http://svn.haxx.se/tsvn/archive-2005-06/1030.shtml (accessed 9/11/2014)
$the_cow = <<EOC;
  $thoughts
   $thoughts       ___
      oo  // \\\\
     (_,\\/ \\_/ \\
       \\ \\_/_\\_/>
       /_/   \\_\\
EOC

```

### turkey
```perl
##
## Turkey!
##
$the_cow = <<EOC;
  $thoughts                                  ,+*^^*+___+++_
   $thoughts                           ,*^^^^              )
    $thoughts                       _+*                     ^**+_
     $thoughts                    +^       _ _++*+_+++_,         )
              _+^^*+_    (     ,+*^ ^          \\+_        )
             {       )  (    ,(    ,_+--+--,      ^)      ^\\
            { (\@)    } f   ,(  ,+-^ __*_*_  ^^\\_   ^\\       )
           {:;-/    (_+*-+^^^^^+*+*<_ _++_)_    )    )      /
          ( /  (    (        ,___    ^*+_+* )   <    <      \\
           U _/     )    *--<  ) ^\\-----++__)   )    )       )
            (      )  _(^)^^))  )  )\\^^^^^))^*+/    /       /
          (      /  (_))_^)) )  )  ))^^^^^))^^^)__/     +^^
         (     ,/    (^))^))  )  ) ))^^^^^^^))^^)       _)
          *+__+*       (_))^)  ) ) ))^^^^^^))^^^^^)____*^
          \\             \\_)^)_)) ))^^^^^^^^^^))^^^^)
           (_             ^\\__^^^^^^^^^^^^))^^^^^^^)
             ^\\___            ^\\__^^^^^^))^^^^^^^^)\\\\
                  ^^^^^\\uuu/^^\\uuu/^^^^\\^\\^\\^\\^\\^\\^\\^\\
                     ___) >____) >___   ^\\_\\_\\_\\_\\_\\_\\)
                    ^^^//\\\\_^^//\\\\_^       ^(\\_\\_\\_\\)
                      ^^^ ^^ ^^^ ^
EOC

```

### turtle
```perl
##
## A mysterious turtle...
##
$the_cow = <<EOC;
    $thoughts                                  ___-------___
     $thoughts                             _-~~             ~~-_
      $thoughts                         _-~                    /~-_
             /^\\__/^\\         /~  \\                   /    \\
           /|  $eye|| $eye|        /      \\_______________/        \\
          | |___||__|      /       /                \\          \\
          |          \\    /      /                    \\          \\
          |   (_______) /______/                        \\_________ \\
          |      $tongue / /         \\                      /            \\
           \\         \\^\\\\         \\                  /               \\     /
             \\         ||           \\______________/      _-_       //\\__//
               \\       ||------_-~~-_ ------------- \\ --/~   ~\\    || __/
                 ~-----||====/~     |==================|       |/~~~~~
                  (_(__/  ./     /                    \\_\\      \\.
                         (_(___/                         \\_____)_)
EOC

```

### tux-big
```perl
# Tux the Penguin (large version)
#  seen when connected to irc.uslug.org
$the_cow = <<EOC;
       $thoughts
        $thoughts          .88888888:.
         $thoughts        88888888.88888.
               .8888888888888888.
               888888888888888888
               88' _`88'_  `88888
               88 88 88 88  88888
               88_88_::_88_:88888
               88:::,::,:::::8888
               88`:::::::::'`8888
              .88  `::::'    8:88.
             8888            `8:888.
           .8888'             `888888.
          .8888:..  .::.  ...:'8888888:.
         .8888.'     :'     `'::`88:88888
        .8888        '         `.888:8888.
       888:8         .           888:88888
     .888:88        .:           888:88888:   
     8888888.       ::           88:888888
     `.::.888.      ::          .88888888
    .::::::.888.    ::         :::`8888'.:.
   ::::::::::.888   '         .::::::::::::
   ::::::::::::.8    '      .:8::::::::::::.
  .::::::::::::::.        .:888:::::::::::::
  :::::::::::::::88:.__..:88888:::::::::::'
   `'.:::::::::::88888888888.88:::::::::'
         `':::_:' -- '' -'-' `':_::::'`
EOC

```

### tux
```perl
##
## TuX
## (c) pborys@p-soft.silesia.linux.org.pl 
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
        .--.
       |$eye_$eye |
       |:_/ |
      //   \\ \\
     (|     | )
    /'\\_   _/`\\
    \\___)=(___/

EOC

```

### tweety-bird
```perl
# Tweety bird
#  from http://pastebin.com/isRcSy01
$the_cow = <<EOC;
    $thoughts
     $thoughts
      $thoughts
                    ___
                _.-'   ```'--.._    
              .'                `-._ 
             /                      `.     
            /                         `.  
           /                            `.  
          :       (                       \\   
          |    (   \\_                  )   `.  
          |     \\__/ '.               /  )  ;  
          |   (___:    \\            _/__/   ;  
          :       | _  ;          .'   |__) :  
           :      |` \\ |         /     /   /  
            \\     |_  ;|        /`\\   /   / 
             \\    ; ) :|       ;_  ; /   /  
              \\_  .-''-.       | ) :/   /  
             .-         `      .--.'   /  
            :         _.----._     `  < 
            :       -'........'-       `.
             `.        `''''`           ;
               `'-.__                  ,'
                     ``--.   :'-------'
                         :   :
                        .'   '.
EOC

```

### USA
```perl
# USA flag
#
# from http://chris.com/ascii/index.php?art=objects/flags
$the_cow = <<EOC;
   $thoughts
    $thoughts

  |* * * * * * * * * * OOOOOOOOOOOOOOOOOOOOOOOOO|
  | * * * * * * * * *  :::::::::::::::::::::::::|
  |* * * * * * * * * * OOOOOOOOOOOOOOOOOOOOOOOOO|
  | * * * * * * * * *  :::::::::::::::::::::::::|
  |* * * * * * * * * * OOOOOOOOOOOOOOOOOOOOOOOOO|
  | * * * * * * * * *  :::::::::::::::::::::::::|
  |* * * * * * * * * * OOOOOOOOOOOOOOOOOOOOOOOOO|
  |:::::::::::::::::::::::::::::::::::::::::::::|
  |OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO|
  |:::::::::::::::::::::::::::::::::::::::::::::|
  |OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO|
  |:::::::::::::::::::::::::::::::::::::::::::::|
  |OOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO|

EOC

```

### vader-koala
```perl
##
## Another canonical koala?
##
$the_cow = <<EOC;
   $thoughts
    $thoughts        .
     .---.  //
    Y|$eye $eye|Y// 
   /_(i=i)K/ 
   ~()~*~()~  
    (_)-(_)   

     Darth 
     Vader    
     koala        
EOC

```

### vader
```perl
##
## Cowth Vader, from geordan@csua.berkeley.edu
##
$the_cow = <<EOC;
        $thoughts    ,-^-.
         $thoughts   !$eyeY$eye!
          $thoughts /./=\\.\\______
               ##      $tongue)\\/\\
                ||-----w||
                ||      ||

               Cowth Vader
EOC

```

### weeping-angel
```perl
#
# Weeping Angel
#
# Don't blink!
#
# based on design found at http://shirt.woot.com/derby/entry/73182/dont-blink
#   and http://infinitywave.deviantart.com/art/Don-t-Blink-tee-391963389
#
$the_cow = <<EOC;
       $thoughts
        $thoughts

                                     ...I..
                            :XX:X\$ . .7N..            ..\$\$.. .:~..
                            X:XX.. 8XXI..         ....XX..7..7KKK8.. .
                              N. .XXX,          ..:ZD- ..M.\$K:XN?XX.XN .
                              .. XX\$.            *. .KN7XXX+ -XX,CN.,XX
                        IXX?..                  ...--+..IXX:X:X..-ZN?DX,.
                      .\$XXXXX. .X               ..XX~-7=\$7+IX5\$...+IM+XXXX.
                    .  +....7D=               .7=IX,: 7+..   . ,.  =+XXID..
                    .-\$-.. .    .             .MM-,,..... . ,=7OI.. .,:N7%
          ..17KN. ..XX:XZ.  ..., .            .:.. .     .-IN78IN7=-.,..CMO.                     ..-.
        ..XXXXMO..  ..8X:X= 8D0..               .8 .    .+I:N-X:XXDXXXX.ID..                  ..-XX:XX- .
        .X:DX:XX.     ...8KI78M,                .X......-IM.D.XDXXXXXXX?:. .                .:++7CXXXXXX?.
      . X7XXXXD8 .      .  ?8XXX.....            X+ : ...XX.X.DCO.  .+X-8X,.              . .=?2XXXXXXXXXX.
       .=XXX887+.       .=:+?,:I.XXXXZ.         .7NO.=8+ ...M?.. 8X.\$\$X 7=..              ..=?X:X7++\$NXXXXX:
       =XX:D-...  .:..  .  ...  ...?XXX         .*. X(O)X  :X .X(O)X+X.,.                .*XXX=    . -IXXXD,
     ,-X:XI..     ,.     ...           .       .X..........XX-....=.?N?.O?               .+XXX..       .-XX:X.
    .:ZX8-      . .   ..+\$:,..                .DX..D:\$78. .XX?:XXXXD?XZ...               +DXX,    ..  .. =X:XX
    --\$D   .    ......+XX. .I.                  X-.*XX ..XZXXXXX.XXXXXX 8.              -.XXO.    .:=-7=..:8XX,
  ..:7Z. .      ..  +X:XX.  X..                 :I.........XXX....XX:XX  .            ..7X7Z..  .-II\$%>I?+,-X:X.
   :,7..       , ...8X:XX.  XX                  .X ...  . .. .8XX:,-ZXX.              .7IXXX .  .+ODC8:\$II7:ZXX:
  .,=.      ..?.  ,8X\$XX=   .XX..               ..+..  .-..%77.XXX.\$XX .            . *IXXXX    :78DX8D08DN7-XXX.
  .:..       *-...-XXXXX.   .,XX..                8,...O..VVVVV XX:XX               . ZOX:XI  . =ID0X0XODOX:+\$XX..
  .:        .? . ,IX:X:X,    .:X=.                .\$ .N VV ....VVM+XX.              .=ZOXXX.  . IDO0X0X0D0X:0?XX?.
  ,         ...  :IXXXXXX- .....O.                 7- O...,I.Z.X.X:X,               .7\$:\$\$..   .OCO0X\$%DCOXXD\$7X>.
  -  ...    ?..  =+X:X:XXDX:XX+ .. . .            ., .I. .VVVVV  XX\$..              :.\$XXX     +DCXXX\$+%\$D8ZXXDXO+
  , .*...   .%7...,,,%%0%XXXXXXXXXXO-D\$7J0\$: ..-:.. ...W,XXXXX,-IXX                .:7XXXX    .CD88DCW..\$DX:X:XX?%
 .,.-%,-.   . ?7.  .. ..+%8XX:X:XX:X:X:NO\$DX..,XXX.. . ..XXDXXI\$XX .               .-O:XX     -CDOI8XX*.*DX:X:X7=%
.--.:\$.+ .     .I%       .:7XXXXXXXXXX    . . ....      ..*.-XDXX..               .-:XXX .   .=78CD8XX+  +IXXX>..Z
 ?=:=?.+.        .XX..      ...?8:XXXX-.           .    .:-Z+XX- ....... .       .+,DXX.    .7.,\$88DXX:  .:OD%..%O.
 ?+IO--+           -XX.  .        .-7O.    .              .\$\$.. .7.M.\$XXX:...   .++DZM.   ..I\$  .\$8DX7.,..  .. IXX..
 +??D,\$\$              .X8.       .        .*      .   . ..  ...*XX X,X:X-: ON?...,XIO     .-O\$.   .-. =O,. .  OXXX-.
.:?ID,Z7              .. ....:I-..        .      ...  .+..   :7 XXD. XX,:...+XXX-.,*..    .=\$?..  . .:O8.\$D7 .IXXX?
 ,+78:Z7 . ,*. ..*.           .....  .....D . ?+.     .:+.  -X+.XX7.8XZ.*.  ..+7XX,     ..*+:.  ... ,8:+.-D0..7XXX7.
.,=\$8I88   ,7  :XD-.  ....             +.:,...--|.. ..      7% CXZ..XX,.*.  ....+XX+.   ..:.    .\$O..DO,  8D+.-XXXX,
 :-\$X0XX ..,*. -DX,... 7... .           .X...::::::-O...    .  .. .7XX...  .,XX. 78-+:..        .OZ.-87   CD\$.7XXXX.
.--7O0XI.. :-O:ZM8...-*O.  -.           :% .::-,,::..:,::::--*+I88XXXD....X:X:XX.7XXXXX:        .8%7DO....-8O77DXXX,
 -=78XX.   .**%XX .   %7:..Z .          8. .:X7.::*,* :,,,::-:-*7\$DXX+.,.XXX?I+.7X:X:X.-X:      .-%%8O .  .8ODIDXXX,
 -=7IX8.    .*87.   ..\$O7-I\$..    -.    .... .. ....:,:O:,::-: .7IDXX ..XX..  .:X:... XXXXX.    ..=7+.    .=0D7CXXX:
.-+7IN..          . :..O\$7X,      =.  ..,.  ..   ...   O..7.::. ?.XXX,.  .      ..  .X:XZ08. . .          ..\$D8%XXX:
.-7II.            .+%. :77:.    .:..  . XD  .-  :...  .:. .   ..... ..          .-  ,X7...,X%.:            . DO8XXX,
 ,I* ,       -?:..8XX. . .....   ...   :XX...X .\$  , .?%: ,? ==7X7., .           D..? . .XXXXX:%...         ..%8DXX
 ..      ..  DOX .8XX..-..:==.:  . .   .X%  .X .% .%. XXO DX.D?\$X?               XX.  . 7XI.XX.. :.   ....    ..-?,.
        ... .DXX..8XX .?8-Z\$?..*.* .    X-  .X..7 7\$..XXX.DX+XXXX,   .        . .XX.  .  .  .XX...     .*.  ..  ..
  ....  :,:. IXX-.8XX .-XDXX- .++\$..  . X,  .X.%,.+% .XXX.\$XIDDDD. .=.  .. ...,8XO..        -XXXX ..    *...., .....
   ,,\$. -:*. -DXO+8X\$ .,ZXD7,  .:...  .,X.  .?.X- +-..XXX 7XOXXXXO. ..  .NXXXXXXXXX         87%:XX..    -..I7\$:.,-.
 . .DX,.-I% ..\$XX8DXI   *OI-..  ..    .XX   ..7X8.?,  XXX.\$DX:XXID     .XXXXXXX8.. ..       .+XXXXX+  ..,  \$XX..:?..
 . .8X-.-%8...*DO\$XX7   ......         XX   . XX .I.. XXD 7XX:XX?,     ,..              .8D\$X:XXXXXX   ... \$8X. ,I7.
    I8\$.-OXX..,\$D8XX... .  -,,....     XX.  ..X  .?:. ZXO -XX:X\$,:...                 .... :XX:IXXXX   ... %8X  .\$X.
..  -8D.-O8X...-\$CX+ ..D,  -X7,-,.     XX   . 7. .I7 .7X\$ :X\$XX-.XXI .                .:: ..X:XXXXXX.. ,. .%XX  .CX:
..  .CD.+OCX  . .,. .-.XO. ?8I,IX.  . :X7.  ...   78..8X-.+X+XX, XXD.     .     .:    ..*N= -XXXXXXD  .:. .OXN  .8XI
 .   78*COyX,       .8*XX  =8....   ..\$X..... .   XXI XX .XX.XX..?XX..7.  ?,    ..      ..?XXXXXXXX.  ,:. ,OX*  ,DX\$
 .  .%D7CO7X\$ . .... X7XX   ..        XX . ..   ..XX\$ X8  XX.XX...XXX%8   %-              .. .=\$XX..  ,:. :ZX   :DX%
.... +XO8O7XD ...+.  X8XX.          .:X.  .7-    ,XX\$.D= .XX 8X   DXX\$X:  XD.                         .-,.*OX.  =DX%
 .,. .XD8O\$XX  ,,\$:. OXXX-          .XX.  .D?.  .XXX:-8. +XX DX:   XX7X\$  XX                          .=-IZXD   +DX8 .
 ::.  XXXXZX\$  .,\$D  7XDXI          OX... O8?   .XXX.%8  IXX ZXI   ZX7XD. DX.             ..   .      .=+Z\$X=   7DX%
.-:.  IXXN7X. ..:IX. 7DNK\$        ..XZ  .\$XX- ..OXX8 %Z  7XX.-XX ..+XDXX+.7X-.           :+   .D      .-*\$+X... %DX%
.*,,. .XXDIX, ..??X. IDXXI        .XX   .XX8,  :XXX  %% .=DX..XX.. .XX:X8 .XX.           .=.  .X      ..+??X  ..%DX?
.+*-  .OX7?%. .:%77\$.7DXX         -X7.  XXX+...XXX:. ZO...DX\$.OXX  .+X\$XX,.XX,          ....  ,X       .-8\$.  .-OXX+
 *I*,. .8+ZI  .=D+XX ,8XX.        XX. .8XX%...7XX7...%X   8XX .XXX  .X\$XX\$.XXX..              -X      . -X. ..?+DXX:.
 =O7%:. :-I  ..*O+XX  I8,.       .XI. .XX%.. -XXD   .:X:  .XX* ZXXZ..\$X7XD.:XX7.              ?8        .-...I8\$XXX..
 -X\$O%:    .,..=7-XX. .,..      .*X   .OD.  .:+%..   *XD  .8X- .7XX..XX..II +CO               -.     ..   .,%XODXXX...
 :XOOX-.  ..?..:I:XX..                 .\$,  .:.     ..,:  .I?,. +8?..\$%.  . ..,.            . ,.    .?..   =XX:XXXX
 .XXXD=    ,O. .I.\$M=                    .     .     ...  ......-....,..     ..             ..X=     X .   -DXX8XXX
 .XXX8*.  ,,\$:. .?+XX                                               .                        =N?..   D     :OXD8XXX
  %XXXI.  ,,7\$. +I78N.            ...                                                        ID?.   .I    ..%X8CXX+
  +ICOI   .,:Z- .-XXD.            .,..                                    ..                 7D*..  ..      \$D\$ZDD..
  ,IOO+.  ,::Z7..7\$\$D              ,.. .                                  .@.                -O\$.:  7.     .78OZ8D..
  .888..  ,-:7O+.7778.            .,.  I .     ..      ....     ...:..  . OO..               .--*:. 7.. .   :%XZ8.
.,.:%I..  ,+:?O?.=I7\$.             .   .      ....... . ::    , .*7%..  :.OO=.               ..-+..,?.. .   .I?OI.
    .     ,7:=\$7,,II7.            ..  ..   .. .-    ..  +7.   :  77\$.-..? %%+                 .7\$. +=:,..   .,..
          .+::7Z7.:I+.            ..  :.    .. -* .. \$  I\$ .  -,.I\$7 =  7.XXI                 .*...*,,.... .    .
        . .-*:I\$?..*.              .  *.    .  -7.. .7. I\$.,  :-.7\$7.=  7-\$\$7                ..    .,- .  ..  ...
    . ...  :7:777. .                  -.    +- +7.   7..7\$+ ..?.?77  =  7+777                 .. ..7?I....-   .. .
    = .....,7:-7I-.                   -.    II =I:   :?.?I7   ?.?II. = .I?7::                   ..+--?.. -*,  .=?
    ,,..,. ,**,I?I..              .   ,.   .II :*?.   ?.??I.. ?.:?I. - .I?II?                   . ,:*?-. ==I..=I?.
    .=...+ .*+,-??.*              .   .:.   +7:.-7.  .+,+??:..*:.+7. :. 7777+                   .,7+++7. -+,.-=7*.
    .=...*. :=:,++:.  .           .   .:. .. =+=.-*: .=*=**= .:-.**, ,..*****.                  .,:=*=*.,:*..***=.
    .....-. ,*-,===.              .   .:  .  ,== :-*. .,*,*+*. .-.** :  -****.                   ,:**:*.,:+..*+*:
    .. .... .:-:=== .             .   .:     .*: ,:*.  .*,--*,  -.-***  :****,                   .:---: ,:-..-**,
.   ..    .. ,--:-- .             .    ,     .-: .,-,.  -::--,..:,---:, .----:                   .:=:=. .::.,===.
       :-::. .::::- .             ..   .     .::  ,:-   .::--:, ,,,;:.  .---::                   .,:.: ..:,.::-:.
     ,..,,,.  ,::::                .   .     .,,  ,,:.  .:,:::,..,.:::. .:::::                   .,,,, ..,.,::::
        ,,,,  .,,,,                .   ..     .,  .,:,   ,,:,,,. ,.,,,.  .:,,,                    ..,.. ...,,:,,
    ,,...,,,   ,,,.                    .      .,. .,,,   .,,,,,. ,.,,,.. .,,,,                    .    .,,.,,,,.
     .,......  ...                     .      ...  ,,,   .,,,,,. ...,,,. .,,,,                          ...,.,,.
      ,......  ...                  .  ...     ..  .....  .................,.,.                           ...,.
      .......   .                        .     ... ....   ....... ....... .....                      ...........
      ... ....                          ..     ...  ...   ....... .............                        ........
       .......                          ...     ..  ...    ...... .............                       ........
        ......                          ...     ...  ...   ... ..  ............                         ......
        .....                           ....      .   .      ....  ............                        .... .
        .. ...                       .. . ..           ..       .   . .... .. ..                       ....
           . .                          . .       ..   ...   ...      ... ....                          .  .
           .                              .       ..          .       .  .   ..
             .                           .                    . .    ..    . .                             .
EOC

```

### whale
```perl
# whale
#
# modified from https://www.reddit.com/r/pics/comments/25ji0n/man_face_to_face_with_whale/chi1kdy?context=3
$the_cow = <<EOC;
   $thoughts
    $thoughts
     $thoughts
                '-.
      .---._     \\ \.--'
    /       `-..__)  ,-'
   |    0           /
    \--.__,   .__.,`
     `-.___'._\\_.'

EOC

```

### wizard
```perl
# Wizard
#
$the_cow = <<EOC;
  $thoughts
   $thoughts
                     _____
                   .\'* *.\'
               ___/_*_(_
              / _______ \\
             _\\_)/___\\(_/_
            / _((\\- -/))_ \\
            \\ \\())(-)(()/ /
             ' \\(((()))/ \'
            / \' \\)).))\\ \' \\
           / _ \\ - | - /_  \\
          (   ( .;\'\'\';. .\'  )
          _\\\"__ /    )\\ __\"/_
            \\/  \\   \' /  \\/
             .\'  \'...\' \'  )
              / /  |   \\  \\
             / .   .    .  \\
            /   .      .    \\
           /   /   |    \\    \\
         .\'   /    b     \'.   \'.
     _.-\'    /     Bb      \'-.  \'-_
 _.-\'       |      BBb        \'-.  \'-.
(________mrf\____.dBBBb._________)____)
EOC


```

### wood
```perl
#
#  木
# 木木
#

$the_cow = <<EOC;
  $thoughts
   $thoughts   --木--
       ／｜＼
     ／  ｜  ＼
  --木-- ｜ --木--
  ／｜＼    ／｜＼
／  ｜　＼／  ｜  ＼
    ｜        ｜
EOC

```

### world
```perl
# World
$the_cow = <<EOC;
   $thoughts
    $thoughts
          _,--',   _._.--._____
   .--.--';_'-.', ";_      _.,-'
  .'--'.  _.'    {`'-;_ .-.>.'
        '-:_      )  / `' '=.
          ) >     {_/,     /~)
  snd     |/               `^ .'
EOC

```

### www
```perl
##
## A cow wadvertising the World Wide Web, from lim@csua.berkeley.edu
##
$the_cow = <<EOC;
        $thoughts   ^__^
         $thoughts  ($eyes)\\_______
            (__)\\       )\\/\\
             $tongue ||--WWW |
                ||     ||
EOC

```

### yasuna_01
```perl
##
## やすなちゃん
##  
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
  
            . .: ───:. .
         .／.: .: .: .: .: ヽ
        .:   .:l.:   .: .: .:.
        |.l:..ﾊ.ハ..|ヽ.ﾄ､:: |
        |:l.:/ヽ､_ヽ|_ノV:.:.|
        |:lﾊ.  {j    {j  |:ヽl
        ﾉ:l} ''        ''|:ノヽ／ )
        ヽﾍ:ヽ.､ r---､  ｨﾉ ┬' `／
     γ::ヽ  ｀^Y`TﾇΤ` {__├'`'
     ｀‐< ＼_ ハ |:|  Y
          ヽ_>､|  |:|／|
               /   V   l
             〈        〉
           〈:｀-:';`-´:〉
            .>-:ｧ─--‐r-:ｨ
            /  /     |  |
           /  /      |  |
          /-,/       |--|
         に7         |二|
EOC

```

### yasuna_02
```perl
##
## やすなちゃん
##  
##
          $the_cow = <<EOC;
           $thoughts
            $thoughts
                           _.. .:-―-:. .._
                      .: .: .: .: .: .: .: .: .: 
                   ／ .: .: .: .: .: .: .: .: .: .＼
                 ,'         ,!    ∧           : .: ヽ
                /, .:: :｜./ |.:./ヽ.:iﾍ.: .: .: .: ::.
               ,''|.:: .人/--|':/  ヽ:| ､＿.: : .: .::|
                  |.:: ｲ  ,,=､ﾚ        ゞ=ﾐ､.:|..: .: :|
                  |.:: ｜{{    }}     {{    }}八.: .: :|
                 /.: : /  ゛= "        ゛= "    ;.:r ､:|
                /,.ｲ.:〈                     ,, //' ｝:|
               '  ヽ:: ゝ、        ｰ--┐       //  ノ::.
                    ヾ::.､＞ .    ヽ _ﾉ    ..  ＜¨ｨ.:}~＼
                      `゜ヾ/｀>了、.    v 〔:／|:/  レ'
                        _ . -/: ,K:::>､/: :ﾄ._
                       |: :く_.:|/:〈 /: :}: /~ヽ
              r「「「ｈ,>:|: <: |'::ｿ::<¨.:n｢「「!､
              ゝ＿_ﾉ /: : |::ヽ |::/: ／: :.ﾍ＿_ノ｝
               | ￣ |,': :/: : ヽ:' ／ : :.:| ￣ |:}
EOC

```

### yasuna_03
```perl
##
## ソーニャちゃんおめでとー！ソーニャちゃんおめでとー！
##
$the_cow = <<EOC;
            . .: -----  .
         ／: .: .: .:.: .:＼
        /    ..  . l.: .: .:ヽ
       : .: ,/|-/|:ハ.:|-.ｌ.:
       |: :ノ |/.|/  ヽ|.Vﾊ.:|
       |.::|  =＝     ＝= }.:| 
       |.γ|| ''  ＿_   ''{::ﾊ 
       ﾉノﾊﾘ   ｛   }     ﾉV
       ∨Vvヽ､._  --'_ .イV
            γ:/:{.又 }ﾍヽ 
          ／:〉:V ﾊ.ﾘ〈: ＼ 
        ／ : Vヽ:V// /:V ::＼ 
    rイ: : ／|: :＼Vノ: :|ヽ: ヽ-､
   ｢  ヽ:／  |: o :  o:|   ＼:/ 」
    ー'    ./: : : : : ﾊ      ー' 
          ./::o: : : :o ﾊ 
          /ヽ: : :Λ: : :ﾉ:、 
        〈:::￣￣:::￣:::::〉 
          ＼:__:::::::__:／ 
            |  Τ￣Τ | 
            |  |   |  | 
            |''|   |''| 
EOC

```

### yasuna_03a
```perl
##
## ソーニャちゃんおめでとー！ソーニャちゃんおめでとー！
##
$the_cow = <<EOC;
 $thoughts
  $thoughts
   $thoughts
            . .: -----  .
         ／: .: .: .:.: .:＼
        /    ..  . l.: .: .:ヽ
       : .: ,/|-/|:ハ.:|-.ｌ.:
       |: :ノ |/.|/  ヽ|.Vﾊ.:|
       |.::|  =＝     ＝= }.:| 
       |.γ|| ''  ＿_   ''{::ﾊ 
       ﾉノﾊﾘ   ｛   }     ﾉV
       ∨Vvヽ､._  --'_ .イV
            γ:/:{.又 }ﾍヽ 
          ／:〉:V ﾊ.ﾘ〈: ＼ 
        ／ : Vヽ:V// /:V ::＼ 
    rイ: : ／|: :＼Vノ: :|ヽ: ヽ-､
   ｢  ヽ:／  |: o :  o:|   ＼:/ 」
    ー'    ./: : : : : ﾊ      ー' 
          ./::o: : : :o ﾊ 
          /ヽ: : :Λ: : :ﾉ:、 
        〈:::￣￣:::￣:::::〉 
          ＼:__:::::::__:／ 
            |  Τ￣Τ | 
            |  |   |  | 
            |''|   |''| 
EOC

```

### yasuna_04
```perl
##
## 湿布
##
$the_cow = <<EOC;
            $thoughts
             $thoughts
              $thoughts
                    .  .:----.:  .      
                  ／    .: .: .: .:＼
                 .          .. .: .: ヽ
                /.: :/|:/.:ﾊ::ﾊ : .: .:.  
              ノ.: ./-|/ |/ V- V､.: .:.:|            
               ｜:ノ   _     _   V: .:,:|
                |:}  =＝     ＝= |:l､:.:|
                |:ﾉ''    ＿_   ''|:| }::|
               八:ヽ.   V_ 丿   .|ﾉｲ: :八
                 ヽ/≧=-z:-:r:=≦l:ﾉ|:／
                    ／/ ﾚﾇﾘ: 〉 ＼
                   / :〉|/l/:< : ハ  
                  /:}:{:|:/:/ : : :.     
                 /: { : ': /: : {: :|            
                 {: : :ヽ:/: : : : :}          
                /} :}:o : : o: {: : ﾊ                
                {: :ﾘ: : : : : |: : }  
EOC

```

### yasuna_05
```perl
##
## ダーツの才能
##
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
       ／ .: .: .: .: .: .: .:  .: .: . ＼
     ./   .: .: .: .: .: .: .:  .: .: .: .ヽ
     /          /  . ..l..  ヽ.: .: .: .: .:.
    ,    .. .: /  .| : ハ: .|  ＼.: .: .: .: .
    |.: .:.l.:/  ヽ|.:/  ､ .|.ノ ＼ .l:.: .: |
    |.: .:.|:/.ｨ≠ﾐ|:/    ＼| ィ≠ミ､|.:.: .:|
    |.: .: ノ /Y::::ヽ       Y::::ヽヽ＼ .: ｜
   /:.: /^|:|{.{:::::}       {:::::}.} |:|ヽ:､
 ノ:ノ: { |:| Ｕうーソ       うーソ  |:| }ヽ:＼
    | : ヽ|.|  '' ￣           ￣ ''U|:| /:|
     :: ::人|                        |人::ﾘ
     Vﾊ:: :: \                     /::  ﾊ/
      \|ヽ:: ::ヽ､     --      ,イ::／|／
          ＼| ヽ:≧=r-r---r-r=≦:ノ|／
             . :´:.ヽ二二.ノ: :｀: .
            ／: : :  ／ハ＼: : : : ＼
EOC

```

### yasuna_06
```perl
##
## 策士
##  
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
 
              ,.:￣￣￣￣:.､
            ／ .: .: .: .: .:＼
          ／   __ l  __    .: .ヽ
        ／:  ./ＶＶＶ＼:ﾄ､`.: .:.
   (＼  ￣/.:ｲ.ｨ=ﾐ` ´ｨ=ﾐ､ヽ,: .:|
   {ﾐと^ヘl.:ﾉ{ぅｿ,  ぅｿ}|:|^ヽ:|
    ヽ〃: ﾚ｛    __ -､ " |:| ﾉ::|
      ＼: :ﾊ＼  {    }  ,ﾚイ:::八
        ＼ :V.:>:ニニﾆ:＜ＶＶＶ
          ＼ :v:〈|父 /:|:ﾊ┐
            ヾ{:｢|/:|/ <:/:ﾉ＼
              {:\|::/:／:{: :ヽ
              {: : : :`: /> : :>
             / : ﾟ : : ﾟ : Y: :/
            /: : : : : : : |ヽ/
           〈: :ﾟ: ﾊ : :ﾟ: ∟ｺ
           /::---':::------く 
EOC

```

### yasuna_07
```perl
##
## ごぼう
##  
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
     $thoughts
               ＿＿＿＿＿
           .:´.: .: . : :. `  、
     ..: ／.: .: .: . : .: .:   ＼
    .::／:::       ﾉ   /､         ＼
   ..:/.: ::.:|＿／::|:/  ＼:__|:  .\
 .:: :::: :::/|／｀ヽ|/    '＼:ﾄ、:  .
 .:::|.:: ::/:ｨf于ミ     .ィ≠ﾐ､Ｖ: :. .
..:::|.:::ノ::{{:::}       {:::}}{: |＼|
..:::::::_::|::うﾆソ       う:ソＶ: |
.::: /.:/ |:|:ヽヽ       ｀      }: |
.:::/ｲ:{  |:|:    ／￣￣ ｧ      ﾉ  :|
 ..::|.ゝ,ヽ|:   /      /     ／:::八
 .:::Ｖ:::::＞:._ヽ、 ./__ .イ:ﾊ:／
  ..::＼|＼:斗:ｰrﾍ`ア又＜Ｖ|／
   ..::::／⌒: :|:ＶＶ{ヽ:＼
      .:/.: :|::l::ﾍ}/\|:}:.＼
    ..::｢.: :|::＞:Ｖ//|〈:.}.}
  ...::/.:: :|::＼: Ｖ/| / :}:.┐
 ...::/.::::rｰ::::＼:Ｖ|/〈::::.ヽ
..:::/.::::ｲ::::::: ＼ Y::ヽ:::::.＼ 
EOC

```

### yasuna_08
```perl
#
# ごぼう2
#

$the_cow << EOC;
  $thoughts
   $thoughts
    $thoughts
          ,.:──‐-:.,
        ／:.           ＼
      ／:. :. :. }:. :. :.ヽ
     .: :. :. }.:/＼.:|,:. :ﾍ
     |:.:. :. /Ｖノ ヽﾄ＼:|､ﾍ
     |:.:. /Ｖ_ﾆ    ﾆ＿_ {::ﾍ
     |:.ﾍ .| ΓT      | |Ｖ.＼
     |:{ |:|.l｜      | |八:ー
     ハ:`:Ｖ､l｜∠二l.|.ｲ:ﾊ:ﾉ
      _Ｖ＼;＞=r rr r=＜ﾊ／   ＿＿
     |ざ |ﾍ :{Ｖ/V:}:＼      |ご  |
    {ﾐ}く{)}:＞Ｖ/< :  ＞-:-'{}ぼ{ﾐ}
     |ろ_|:ﾉ:＼:Y / }ﾐ : : : |  うY
          ﾉ :o: : :oj `ー─-´ ￣￣
         / : : : : :{
        /: : o : :o:ﾍ
      〈 : : : /\: : 〉
       /::ー──'::ー‐ﾍ
     〈:::::::::::::::::〉
       ￣|￣｢￣|￣|￣
         |  |  |  | 
EOC

```

### yasuna_09
```perl
#
# ちびきゃら
#

$the_cow = <<EOC;
  $thoughts
   $thoughts
    $thoughts
    
           ____
       ,: .: .: :.ヽ
     ,'       /\   ｉ
     {: .:ﾉﾚﾍ/  Viﾍ:}
    .{,､〈 Ｏ   Ｏ{.:.
    ノヽ\!"       }.:ﾊ
      Ｗﾊw=-､へ,ｬ<,V'      
         /ﾍ }{./\
        ;: i:V:!;}
        |:｜: :｜}
        |:|:｡: ｡l}
        >-'-ﾟ-'`ﾟu
        ｰi-i～i-i~
         |.|  |.|
         |-|  |-|
         ヒｺ  ヒｺ 
EOC

```

### yasuna_10
```perl
#
# 何でそういうときだけ凄そうなの！
#

$the_cow = <<EOC;
  $thoughts
   $thoughts
    $thoughts
             ＿＿＿＿
       ＜ :: :: :: :: `丶､
       ／   _, ｨ:ﾊ ､＿: ::＼
     ∠:: :/ |/|/ \/  \/:: |
rヘn  /:\/ c=＝.::.＝=っ\/ |  rvへ
ヽ／＼i:｜   ┌──┐   i::|／＼ノ
  ＼::|(||   |:::::|    ||)|::／
    ＼|人|.、|:::::| .ｨ|ﾉ:八／
      ＼\/\/>|:::::|<\/\/／
        ＼ :::>TﾇT<::: ／
          Y : ＼W／ : Y 
EOC

```

### yasuna_11
```perl
#
# きゅーっ！
#

$the_cow = <<EOC;
  $thoughts
   $thoughts          . .: -ーー― :._
    $thoughts       ／.: .: .: .:     ＞  r⌒ヽ
           / .:         ｜.､.:＼  ﾉ ノ
          .: .: .:|＼  |斗ﾍﾄ.:.:Ｖ  /
          |: .: /\|ノ＼| ／ Ｖ::Ｎ./
          |: .:/ c─-        Ｙ:| /
          |:ﾊ:{``   ,  --┐  人V /
          ﾉ:L＼>   く_,￣┘／  ＼
   /⌒￣￣￣|￣￣＞--r-rｭ＜|   ／
   L_,vー─-|    ､ }  ＶYﾊ   Y
             ￣￣Ｖ  ｜/∧   ﾍ
                  {   |//∧  ﾍ
                  {    ＼//   ﾍ
                  {            ＼
                  ｝             >
EOC

```

### yasuna_12
```perl
#
# からあげ
#

$the_cow = <<EOC;
       $thoughts
        $thoughts
         $thoughts        .:  ￣￣￣￣:.丶､
               ／.: .: .: .: .: .: ＼
              /    ／|    /\.:| .: :.
             / .:|乂 |/{:/ _乂/\ .:.:|
           ノ.:\/ｨ庁ﾐ` \/ｨ庁ﾐx  \/:.:|
             |:}{弋.ﾉ    弋ノ } /.:.:|
             ﾚ:ﾘ''          '' ｜:ハ:＼
             {人       ,、    ,｜/ノ:厂 
EOC

```

### yasuna_13
```perl
#
# 転んでも泣かない！
#

$the_cow = <<EOC;
       $thoughts
        $thoughts    ｡
         $thoughts       ＿＿＿__{_    o
      ○   ヽ￣    .: .: .: ｀丶＿_
           ／ .: .／;|.:/|::∧.:＼        О
     (ヽn∠ .::|∠二:|:/:|ン-:∨:「ﾚ^L
     ζ, ヘ /::(___)|/:(￣￣ )|.:/、  ζ
     `く: :/_:ﾉ(_) ＿＿ ￣(_) |:/:`ｰ:/
  。    ＼|＼人   |/   `⌒ヽ ｜/ : :/  o
      __┌   ＼`ヘ|/ヽ/ヽ／^ヽ/／/:/／/
      ＼                        /:／ / 
EOC

```

### yasuna_14
```perl
#
# くっ、くぅー！
#
$the_cow = <<EOC;
       $thoughts
        $thoughts
         $thoughts
                 .:-────-:.  .
             .: .: .: .: .: .: .: :.
          ／.: .: .: .: .: .: .: .:  ＼
         .: .:          /:  |        :.
        .: .: : :  |.:./ |.: ハ.:.|::|.: :＼
       /.: .: .: .:|.:/ u|.:/ u ､:|::|.:｜―`
      /.: .: .:|:,|.＼._ｨ.:/ ､_／｜∨,::|
     /.: .: .: |:/ィ≠ミ |/  ィ=ミ､ ∨::｜
    /..: ,--|:｜  {んi:i}     ri:i}} ﾊ::|
   /.ノ.:/へ|:|.   ∨:タ...::.ヾ:タ  .:.:､
   ／:: :ﾊ (|:| u ''       '    ''  {:|ヽ:＼
     {: :＼_|:| ｕ   __          u ﾉ:｜
     ∨ﾊ:ﾊ:ヽ.|､   （- `ｰｧ      ..ｲ::/ﾉ
       ,:＜:￣/|､＞:._￣..:-=≦::ﾊ:／
      /: ヽ::/:| ＼_ィ .ハ＞:、
     」: : :く:｜／{;;}∨: }::ﾊ
    /:＼ : }/￣`Yヽ:∥:／: /:「Y二ヽ
   / : : : /  ￣}-':/::〉: }:/Y{─ }
  /: : : :/  .二ﾌ::/::/: : ﾘ::ﾊ{-- ﾉ
./へ──‐ﾊ  ,-ｲ :/::/ : :ﾑ:-{､_エノ
{: : : :.ヽ>イ:|:/::ノ: :/ : {{ ／ﾉ 
EOC

```

### yasuna_16
```perl
$the_cow = <<EOC;
   $thoughts
    $thoughts
         ..: ￣￣￣￣: :.
       ／::  /｜.:/ |.: .:＼
      ,  /｜/  |./  |.ﾊ.: .:ヽ
    ./.:ｲ__ノ   ヽ､___∨.: .:.
   ./: .:≡≡     ≡≡.|.: .:｜
   /ノ|/} }.      } } |:ﾊ:.:｜
     .ヽ{,{ -~~~- {,{｜:/ﾉ:从
      ∨v､＞z-r-x-:r＜/ﾚﾚへ 
EOC

```

### yasuna_17
```perl
#
# さっそく試してみよう 道具持ってないから作るしかないかな
#  

$the_cow = <<EOC;
   $thoughts
    $thoughts
     $thoughts                 ____
                 .: :<::. ::.>: :.
               ／:: ::. :. ::. ::`:、
               `::. ::.ィ:.i::.、::.ヽ
             /'      ./|..ﾄ.}V.. .. ﾊ
            '.. .. ./L/｜:| 一V::. ::１
            i::. ::/}/` V:| V Vﾄ::. ::i
            |::. :/Y芋ミV!Y 芋ミ|::. .|
            ,::. ハ {::}  V {::}}:r,:代
            /::. :}  つﾉ    つﾉ｜:レ:}ゝ  ヽ
              V::八    r一 ┐   ｨ!::.:ﾘ      }
       ｛r     ＼ﾊ:＞- .一-'.s<:ハ}ヽ}   __ノ ﾉ
        弋二一   ヽ:{＞}_ノ  / ゝ､
                ｡＜   〈ﾊ〉  {    `、
              ／     i       `､.    `、
            ／    フ^|   　   ',ﾞ、   `、
           く   ／   |         ', ﾞ、y ヽ
           tゝ_r     r          ',  ><一'
                    /  ゞ＿      '
                   /      一      `
EOC

```

### yasuna_18
```perl
#
# ま、ありがちな言い訳だよね
#  

$the_cow = <<EOC;
   $thoughts
    $thoughts
                      ,:二二二二:. .,
                   ／.／＿＿＿_  ＼.:＼
                  /. /／.: .: .:＼  : .:＼
                 /.: .: .:/｜:/\ .:＼}.: .:.
                .: |.:/一/ |:/ 一.:}: .: .:｜
                |.:|ノ |/_｜/ _  \/ﾍ: .: .:|
                |.: ｜= ＝    ＝＝= \/}: .:|
                |:: ﾘ''           '' /:/､.:|
               ノ:|:人    一一 ､    /:/ ﾉ.:|
                , ┴＜＼  {     ｝ ,{:/イ::八
               /_..   ＼` ー┬一r＜:八八／
               ／  T＼   `＜}ゞ=彡'⌒＼＼_>
              /___ |  >､    ｀''＼   ｜
             /ﾆ}::\/／  ＼       ｜  ｜
          　{ﾆﾉ:: /''＼ | `|r--ｯ＜|_／|
           /__   V    ｝|  》=《      |
           ＼ ＼/｀一ﾍノ|  { 6 }     ｛ 
             ￣        ｢   ゞ= '      }
                      ﾉ               〉 
EOC

```

### yasuna_19
```perl
#
# やすなちゃんのまんまるお目目
#

$the_cow = <<EOC;
  $thoughts
   $thoughts

:. :.孑|:/仔:./  ＼:.| V｜:. ﾄ:. :.
:. :/  |/  |:/     ヽ|   \/:!\/:.:
:. / ,ィf芋ミ     ィf芋｀:V  .\/.
:./ ,' :'::::ﾊ      ,':::::ﾊ ヽ /:.
:t  { {k)::::!     !k)::::!  },'.:
:ﾊ    弋 一ソ      弋 一 ｿ ,: ::
:.{      ￣    ,       ￣  ; :./
:.| ''                  '' |:./
:.ﾄ､      ` ､      ノ     ﾉ!:/ノ
ﾄ､!:＞ ､.     一  '   .,＜:|/::.
:: :: :: ::>z-一-z<:: :: :: :: :.
V|＼:/}ﾍ/  `ー又ー' \/}ノ{／|:／
  ,z'￣ ﾍ   /{ .ﾄ､  /￣  ヽ
／      /\./x 一 ﾐ./       ＼ 
EOC

```

### yasuna_20
```perl
#
# yasuna_20.cow - もしかしたら新種かも！
#

$the_cow = <<EOC;
  $thoughts
   $thoughts
    $thoughts            ________
             .:          :｀丶
           /.:   ｛ :｜､  .: .:＼
          /   |.: /\.:|ﾉ＼.} .: :.
         .: .:/\乂  ＼ｨ=ミV.:}.: |
         |.:\/ ｨ=ﾐ    ﾋソ｝V:|.:｜
         |.:ﾊ{ ﾋソ '    ''｜:|ヽ｜
         |.: ﾊ''          ｜:ﾉノ:＼
        丿.:|人    ⌒ヽ    ｲ::\/ ￣
    /^^ﾍ  \/Vv:＞=rr::rr＜vV\/
  ｛   ﾉ    ノ   \/ヌ\／ ＼
    ＼  ＼,く  }   |:|   V ＼
      ＼     >ィ   |:|   ｝  ﾉ
        ＼／  ﾉ    |:|   }-く ＼
             /      V     \  ＼  ＼ 
EOC

```

### ymd_udon
```perl
##
## 山田うどん
##
$the_cow = <<EOC;
   $thoughts
    $thoughts
  
             _ - ￣ - _
           _-_＿＿＿＿_- _
         ￣ｌ  ●   ●  l￣
            ヽ､_ ⌒ _ノ
         _ -‐ニ ￣ ニ‐- _
  /⌒ ‐ﾆ‐ ￣   /    \ ￣ ‐ﾆ‐⌒ヽ
 ヽ､_ノ       └-ｕ‐┘      ヽ､_ノ
EOC

```

### zen-noh-milk
```perl
$the_cow = <<EOC;
  $thoughts
   $thoughts
    $thoughts

     iﾆﾆi
    /   /ヽ
   ｜農｜｜
   ｜協｜｜
   ｜牛｜｜＿
 ／｜乳｜｜／
 ￣￣￣￣￣
EOC

```
