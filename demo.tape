# VHS tape for vergeos-tui — generate with:  vhs demo.tape
# Read-only tour: navigate, search, sort, detail, tabs, help. No mutations.

Output demo.gif

Set Shell bash
Set FontSize 15
Set Width 1220
Set Height 720
Set Padding 18
Set Theme "Catppuccin Mocha"
Set TypingSpeed 60ms

Hide
Type "clear" Enter
Show

Type "./vtui"
Sleep 500ms
Enter
Sleep 3s

# navigate the VM list
Down@250ms 6
Sleep 3s

# live search
Type "/"
Sleep 400ms
Type "k3s"
Sleep 1.8s
Escape
Sleep 3s

# sort by NAME (col 2), then reverse
Type "2"
Sleep 1.3s
Type "2"
Sleep 1.3s
Type "2"
Sleep 3s

# open VM detail (drives + NICs)
Enter
Sleep 2.5s
Escape
Sleep 3s

# Networks tab + drill-in detail (type, CIDR, DHCP, DNS)
Tab
Sleep 1.5s
Enter
Sleep 2.5s
Escape
Sleep 3s

# Tenants tab + drill-in detail (state, UI IP, UUID)
Tab
Sleep 1.5s
Enter
Sleep 2.5s
Escape
Sleep 3s

# Storage tab — usage bars in the list, then the big detail bar
Tab
Sleep 2s
Enter
Sleep 2.5s
Escape
Sleep 3s

# auto-refresh countdown (watch "next in Ns" tick down), then full help
Type "a"
Sleep 4.5s
Type "a"
Sleep 400ms
Type "?"
Sleep 2.5s
Type "?"
Sleep 3s

Type "q"
Sleep 800ms
