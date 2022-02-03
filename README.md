# woogles-game-gifs
Code for creating GIF replays of games played on Woogles.io.
## Example Output
<img src="https://github.com/reyacd/woogles-game-gifs/blob/main/data/example-GdTkgTga.gif" width="400" height="400"/>

# Requirements
Go is required to run the gif generator. To install, reference https://go.dev/doc/install.

# Running the code
1. Clone this repository.
2. Change to the repository directory.
3. Run **go mod tidy** to download the required packages.
4. Run **go run . \<Woogles game ID\>** to run the gif generator.

# Finding a Woogles game ID
The Woogles game ID is the string of letters at the end of a Woogles game URL e.g. "GdTkgTga" in https://woogles.io/game/GdTkgTga.
