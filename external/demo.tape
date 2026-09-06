# Gogit README demo
# Before recording: git switch main && go build -o gogit .
# Then run from the Gogit repository root with: vhs external/demo.tape

Require git

Output external/gif/demo.gif

Set Shell "bash"
Set Width 1000
Set Height 430
Set Padding 28
Set FontSize 24
Set LineHeight 1.1
Set TypingSpeed 75ms
Set WindowBar Colorful
Set WindowBarSize 36
Set BorderRadius 10
Set Theme { "name": "Gogit Demo", "background": "#0d1117", "foreground": "#f05032", "selection": "#264f78", "cursor": "#f0f6fc", "cursorAccent": "#0d1117", "black": "#161b22", "brightBlack": "#6e7681", "red": "#ff7b72", "brightRed": "#ffa198", "green": "#3fb950", "brightGreen": "#56d364", "yellow": "#d29922", "brightYellow": "#e3b341", "blue": "#58a6ff", "brightBlue": "#79c0ff", "magenta": "#bc8cff", "brightMagenta": "#d2a8ff", "cyan": "#39c5cf", "brightCyan": "#56d4dd", "white": "#b1bac4", "brightWhite": "#f0f6fc" }

# Start the prebuilt Gogit binary. Clear removes VHS's setup prompt from the GIF.
Hide
Type "clear"
Enter
Sleep 1s
Type "./gogit"
Enter
Sleep 3s
Show
Sleep 1s

# Type git branch, then use Gogit to choose its option.
Type "git branch "
Sleep 1500ms

# Select --show-current, visit the options below it, then return to it.
Down
Sleep 800ms
Down@500ms 3
Sleep 800ms
Up@500ms 3
Sleep 1s

# Accept the suggestion and run the completed Git command.
Tab
Sleep 1s
Enter
Sleep 4s
