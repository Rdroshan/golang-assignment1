package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "os"
)

type Actor struct {
    URL    string `json:"url"`
    Name   string `json:"name"`
    Movies []struct {
        Name string `json:"name"`
        URL  string `json:"url"`
        Role string `json:"role"`
    } `json:"movies"`
}

type Movie struct {
    URL  string `json:"url"`
    Name string `json:"name"`
    Cast []struct {
        URL  string `json:"url"`
        Name string `json:"name"`
        Role string `json:"role"`
    } `json:"cast"`
}

type PathInfo struct {
    Movie Movie
    currRole string
    castRole string
}

func fetchActorData(actorURL string) (*Actor, error) {
    resp, err := http.Get("https://data.moviebuff.com/" + actorURL)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var actor Actor
    err = json.NewDecoder(resp.Body).Decode(&actor)
    if err != nil {
        return nil, err
    }

    return &actor, nil
}

func fetchMovieData(movieURL string) (*Movie, error) {
    for {

        resp, err := http.Get("https://data.moviebuff.com/" + movieURL)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode == http.StatusOK {
			var movie Movie
			err = json.NewDecoder(resp.Body).Decode(&movie)
			resp.Body.Close()
			if err != nil {
				return nil, err
			}
			return &movie, nil
	}
    resp.Body.Close()
    return nil, nil
    }
}


func findDegreesOfSeparationBFS(startActorURL, endActorURL string) (int, map[string]map[string]PathInfo, error) {
    visited := make(map[string]bool)
    pathMap := make(map[string]map[string]PathInfo)

    degrees, _, err := bfs(startActorURL, endActorURL, visited, pathMap)
    if err != nil {
        return -1, nil, err
    }

    return degrees, pathMap, nil
}


func bfs(startActorURL, endActorURL string, visited map[string]bool, pathMap map[string]map[string]PathInfo) (int, []string, error) {
    queue := []string{startActorURL}
    degrees := 0
    fmt.Printf("Finding connections")
    for len(queue) > 0 {
        levelSize := len(queue)
        degrees++

        for i := 0; i < levelSize; i++ {
            currentActorURL := queue[0]
            queue = queue[1:]

            if currentActorURL == endActorURL {
                return degrees, nil, nil // Actors are connected
            }

            if visited[currentActorURL] {
                continue // Skip already visited actors
            }

            visited[currentActorURL] = true

            actor, err := fetchActorData(currentActorURL)
            if err != nil {
                return -1, nil, err
            }

            for _, movie := range actor.Movies {
                movieData, err := fetchMovieData(movie.URL)
                if movieData == nil {
                    continue // some apis will give 403, Hence skipping them
                }
                if err != nil {
                    return -1, nil, err
                }

                for _, castMember := range movieData.Cast {
                    if !visited[castMember.URL] {
                        queue = append(queue, castMember.URL)
                        if _, exists := pathMap[castMember.URL]; !exists {
                            pathMap[castMember.URL] = make(map[string]PathInfo)
                        }
                        pathInfo := PathInfo{
                            Movie: *movieData,
                            currRole: castMember.Role,
                            castRole: movie.Role,
                        }
                        pathMap[castMember.URL][currentActorURL] = pathInfo
                        fmt.Printf(".")
                        // fmt.Printf("cast actor %s\n", castMember.URL)
                        // printPathMap(pathMap[castMember.URL])

                        if castMember.URL == endActorURL {
                            return degrees, nil, nil
                        }
                    }
                }
            }
        }
    }

    return -1, nil, nil
}

func printPathMap(p map[string]PathInfo) {
    for targetActorURL, pathInfo := range p {
        fmt.Printf("current: %s\n", targetActorURL)
        fmt.Printf("Movie: %s\n", pathInfo.Movie.Name)
        fmt.Printf("currRole: %s\n", pathInfo.currRole)
        fmt.Printf("castRole: %s\n", pathInfo.castRole)
    }
}


func printOutput(startActorURL, endActorURL string, pathMap map[string]map[string]PathInfo, degrees int) ([]string, error) {
    var output []string
    currentActorURL := endActorURL
    fmt.Printf("\n")
    for currentActorURL != startActorURL {
        actorToPathInfo := pathMap[currentActorURL]

        var targetKey string

        for key, _ := range actorToPathInfo {
            targetKey = key
            break
        }
        pathInfo := actorToPathInfo[targetKey]
        output = append(output, fmt.Sprintf("Movie: %s\n%s: %s\n%s: %s\n", pathInfo.Movie.Name, pathInfo.castRole, targetKey, pathInfo.currRole, currentActorURL))

        currentActorURL = targetKey
    }
    reverse(output)

    fmt.Printf("Degrees: %d\n", degrees)
    for _, i := range output {
        fmt.Printf(i)
    }
    return output, nil
}

func reverse(slice []string) {
    for i := 0; i < len(slice)/2; i++ {
        j := len(slice) - i - 1
        slice[i], slice[j] = slice[j], slice[i]
    }
}


func main() {
    if len(os.Args) != 3 {
        fmt.Println("Usage: degrees <start_actor_url> <end_actor_url>")
        return
    }

    startActorURL := os.Args[1]
    endActorURL := os.Args[2]

    degrees, path, err := findDegreesOfSeparationBFS(startActorURL, endActorURL)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    if degrees == -1 {
        fmt.Println("No connection found between the actors.")
    } else {
        printOutput(startActorURL, endActorURL, path, degrees)
    }
}
