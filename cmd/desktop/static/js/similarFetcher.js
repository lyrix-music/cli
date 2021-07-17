let currentTrack = ""
let currentArtist = ""
let currentPlayers = []
let hexValues = ["0","1","2","3","4","5","6","7","8","9","a","b","c","d","e"];




function getSongHTML(track, artist){
    let encodeSearchString = encodeURI(track + " " + artist)
    return `<div class="card mb-3">
        <div class="card-content">
            <div class="media">
                <div class="media-left">
                    <div 
                        data-track="${makeSafeForCSS(track)}"
                        class="image is-48x48 album-art-square">
                    </div>
                </div>
                <div class="media-content">
                    <p class="title is-4">${track}</p>
                    <p class="subtitle is-6">${artist}</p>
                </div>
            </div>
            
        </div>
        <footer class="card-footer">
            <a class="card-footer-item music-service-provider" target="_blank"
                href="https://music.youtube.com/search?q=${encodeSearchString}">
              <span class="icon"><i class="fas fa-record-vinyl"></i></span>
              <span class="is-hidden-touch">Youtube Music</span>
            </a>
            <a class="card-footer-item music-service-provider" target="_blank"
                href="https://open.spotify.com/search/${encodeSearchString}"
              <span class="icon"><i class="fab fa-spotify"></i></span>
               <span class="is-hidden-touch">Spotify</span>
            </a>
            <a class="card-footer-item music-service-provider" target="_blank"
                href="https://www.youtube.com/results?search_query=${encodeSearchString}">
              <span class="icon"><i class="fab fa-youtube"></i></span>
               <span class="is-hidden-touch">Youtube</span>
            </a>
            <a class="card-footer-item music-service-provider" target="_blank"
                href="https://soundcloud.com/search?q=${encodeSearchString}">
              <span class="icon"><i class="fab fa-soundcloud"></i></span>
               <span class="is-hidden-touch">Soundcloud</span>  
            </a>
        </footer>
    </div>`
}

$(document).ready(function() {

    // Check for click events on the navbar burger icon
    $(".navbar-burger").click(function() {

        // Toggle the "is-active" class on both the "navbar-burger" and the "navbar-menu"
        $(".navbar-burger").toggleClass("is-active");
        $(".navbar-menu").toggleClass("is-active");
    });

    let queue_songs_button = $("#queue-similar-songs")
    queue_songs_button.click(function() {
        queue_songs_button.addClass("is-loading");
        $.get("/api/v1/player/queue/similar", function() {
            console.log("Similar songs queued")
            queue_songs_button.removeClass("is-loading");
            queue_songs_button.prop("disabled", true);
            queue_songs_button.text("Queued!");
        })
        console.log("fetching similar songs");
    })


    $.get("/api/v1/user/logged-in", function(data) {
        console.log("Received data from /user/logged-in")
        if (data["logged_in"] === false) {
            console.log("User is not logged in, disabling scrobble switch")
            window.location.replace("/login")
        } else {
            $("#navBarLoginButton").remove()
        }
    }, "json")

    function getSimilarSongs() {
        console.log("Trying to fetch Lyrics")
        $.get("/api/v1/song/similar", function (data) {
            console.log("Received similar songs")
            let similarContainer = $("#similar-songs-container")
            similarContainer.empty()
            data.forEach(function (item) {
                if (item["track"] === "") {
                    return
                }

                similarContainer.append(getSongHTML(item["track"], item["artist"]))

            })
            data.forEach(function(item) {
                setAlbumArtGradient(item["track"])

            })
            queue_songs_button.prop("disabled", false)
            queue_songs_button.text("Queue All");
            queue_songs_button.removeClass("is-loading");
            
            iswebview().then((x) => {
                console.log("Detected webview running", x)
                if (x === false) {
                    return
                }
                console.log("Setting on click ")
                $(".music-service-provider").on('click', function (e) {
                    e.preventDefault();
                    let url = $(this).attr("href")
                    $(this).addClass("is-loading")
                    console.log("clicked on", url)
                    open(url)1
                })
            })


            // $(".lyrics").html(data.replaceAll("\n", "<br>"))
        }, "json")

    }

    function getUpdates() {
        console.log("Trying to get updates on song")
        $.get("/api/v1/updates/song", function (data) {
            // console.log("Received updates")
            let track = data["track"]
            let artist = data["artist"]
            if (track === "") {
                return
            }
            if (track === currentTrack && artist === currentArtist) {
                return
            }

            currentArtist = artist
            currentTrack = track
            coverArtQueue = []

            $(".subtitle").text(`similar to your ${track} by ${artist}`)
            getSimilarSongs()

        }, "json")

    }


    $("#scrobbleSwitch").change(function () {
        setScrobble(this.checked)
    })



    $(".dropdown .button").click(function (){
        let dropdown = $(this).parents('.dropdown');
        dropdown.toggleClass('is-active');

    });

    getUpdates()
    setInterval(getUpdates, 2 * 1000)
});


function setAlbumArtGradient(track, artist) {



    function populate(a) {
        for ( let i = 0; i < 6; i++ ) {
            let x = Math.round( Math.random() * 14 );
            let y = hexValues[x];
            a += y;
        }
        return a;
    }

    let newColor1 = populate('#');
    let newColor2 = populate('#');
    let angle = Math.round( Math.random() * 360 );

    let gradient = "linear-gradient(" + angle + "deg, " + newColor1 + ", " + newColor2 + ")";
    $(`div[data-track="${makeSafeForCSS(track)}"]`).css('background', gradient);

}


function makeSafeForCSS(name) {
    return name.replace(/[^a-z0-9]/g, function(s) {
        var c = s.charCodeAt(0);
        if (c === 32) return '-';
        if (c >= 65 && c <= 90) return '_' + s.toLowerCase();
        return '__' + ('000' + c.toString(16)).slice(-4);
    });
}