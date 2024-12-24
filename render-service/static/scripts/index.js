const media = videojs('media-player', {
    children: [
      'bigPlayButton',
      'controlBar'
    ],
    playbackRates: [1],
    preload: "auto",
    aspectRatio: "16:9",
    fluid: true,
    autoplay: true,
    controls: true
});

const mediaContainer = document.querySelector(".media-container");
const movieGrid = document.getElementById('movie-grid');
const memberBox = document.getElementById("member-box");
const msgBox = document.getElementById("message-box");
const sentBtn = document.getElementById("sent-message");
const inputText = document.getElementById("input-text");
const participantNum = document.getElementById("participant-value");
const webAddr = "http://116.96.94.33:3000";
const webSocketAddr = "ws://116.96.94.33:3000/ws"

const memNodes = []

var ws = new WebSocket(webSocketAddr);

ws.onopen = updateParticipantReq;
ws.onclose = function () {
    window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
};

window.onload = function () {
    fetch(webAddr + '/clientBoxData', {
        method: 'GET',
        credentials: 'include',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'Cache-Control': 'must-revalidate,no-cache,no-store',
                "Pragma": "no-cache",
            },
    }).then(function(res) {
        if (res.status == 200) 
            {
                res.json().then(function (data) {
                    if (data.is_owner)
                    {
                        ownerLayout(true);
                        setInterval(updateMovieData, 1000);
                        media.on("pause", updateMovieData);
                        media.on("play", updateMovieData);
                    }
                    else
                        ownerLayout(false);

                    document.getElementById("box-id").innerHTML = "#" + data.box_id;
                })
                .catch(err => alert(err));
            }
        else if (res.status == 401 || res.status == 404)
            window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
        else
            res.text()
                .then(text => alert(text))
                .catch(err => alert(err));
    })
    .catch(err => alert(err))
};

sentBtn.onclick = function () {
    let content = inputText.value;
    if (content == "")
        return;

    appendNewMessage(content);
    inputText.value = "";
    ws.send(JSON.stringify(
        {
            datatype: 0,
            content: content,
        }
    ));
};

document.getElementById("expand").onclick = function () {
    media.requestFullscreen();
}

document.getElementById("power-off").onclick = function () {
    fetch(webAddr + '/delete', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Cache-Control': 'must-revalidate,no-cache,no-store',
            "Pragma": "no-cache",
        }
    }).then(function(res) {
        if (res.status == 200 || res.status == 401 || res.status == 404) 
            {
                window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
                ws.close();
            }
        else
            res.text()
                .then(text => alert(text))
                .catch(err => alert(err));
    })
    .catch(err => alert(err))
}

document.getElementById("leave").onclick = function () {
    fetch(webAddr + '/leave', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Cache-Control': 'must-revalidate,no-cache,no-store',
            "Pragma": "no-cache",
        }
    }).then(function(res) {
        if (res.status == 200 || res.status == 401 || res.status == 404) 
            {
                window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
                updateParticipantReq();
                ws.close();
            }
        else
            res.text()
                .then(text => alert(text))
                .catch(err => alert(err));
    })
    .catch(err => alert(err))
}

function kick(id) {
    fetch(webAddr + '/kick/' + id, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Cache-Control': 'must-revalidate,no-cache,no-store',
            "Pragma": "no-cache",
        }
    }).then(function(res) {
        if (res.status != 200)
            res.text()
                .then(text => alert(text))
                .catch(err => alert(err));
        else
            updateParticipantReq();
    })
    .catch(err => alert(err))
}

function updateMovieData() {
    ws.send(JSON.stringify(
        {
            datatype: 1,
            movie_url: media.src(),
            elapsed: media.currentTime(),
            is_pause: media.paused()
        }
    ));
}

function updateParticipantReq() {
    ws.send(JSON.stringify(
        {
            datatype: 2
        }
    ));
}

ws.onmessage = function(evt) {
    let data = JSON.parse(evt.data);

    if (data.datatype == 0) {
        let sender = data.username;
        let content = data.content;
    
        let msgNode = document.createElement("li");
        msgNode.classList.add("message");
        msgNode.innerHTML = content;
    
        let senders = msgBox.querySelectorAll(".sender");
        if (senders.length == 0 || senders[senders.length - 1].innerHTML != sender) {
            let senderNode = document.createElement("li");
            senderNode.classList.add("sender");
            senderNode.innerHTML = sender;
            msgBox.appendChild(senderNode);
        }
    
        msgBox.appendChild(msgNode);
    } else if (data.datatype == 1) {
        media.currentTime(data.elapsed);

        console.log(data);
        if (media.src() != data.movie_url)
        {
            mediaContainer.classList.remove("d-none");
            movieGrid.classList.add("d-none");
            media.src(data.movie_url, type='application/x-mpegURL');
            media.play();
        }
        else if (data.movie_url == "")
        {
            mediaContainer.classList.add("d-none");
            movieGrid.classList.remove("d-none");
        }

        if (data.is_pause)
            media.pause();
        else 
            media.play();
    } else if (data.datatype == 2) {
        participantNum.innerHTML = data.box_user_num;
        fetch(webAddr + '/box/' + data.box_id + '/exists', {
            method: 'GET',
            credentials: 'include',
            headers: {
                'Cache-Control': 'must-revalidate,no-cache,no-store',
                "Pragma": "no-cache",
            }
        }).then(function(res) {
            if (res.status != 200)
                res.text()
                    .then(text => alert(text))
                    .catch(err => alert(err));
            else {
                res.json().then(function (exists) {
                    if (exists.value)
                        refreshMembers(data);
                    else 
                        window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
                });
            }
        })
        .catch(err => alert(err))
    }
};

function ownerLayout(isActive) {
    if (isActive) {
        document.getElementById("request-pause").classList.add("d-none");
        document.getElementById("leave").classList.add("d-none");
        document.getElementById("power-off").classList.remove("d-none");
        media.controls(true);
        loadMovies();
    }
    else {
        document.getElementById("request-pause").classList.remove("d-none");
        document.getElementById("leave").classList.remove("d-none");
        document.getElementById("power-off").classList.add("d-none");
        media.controls(false);

        mediaContainer.classList.add("d-none");
        movieGrid.classList.remove("d-none");
    }
}

function activateTab(tabName) {
    const tabContents = document.querySelectorAll('.tab-content');
    tabContents.forEach(content => {
      content.classList.remove('active');
    });
  
    const activeTabContent = document.querySelector(`.tab-content[name="${tabName}"]`);
    if (activeTabContent) {
      activeTabContent.classList.add('active');
    }
    
    const tabs = document.querySelectorAll('.tab');
    tabs.forEach(tab => {
      tab.classList.remove('active');
    });
  
    const activeTab = document.querySelector(`.tab[data-name="${tabName}"]`);
    if (activeTab) {
      activeTab.classList.add('active');
    }
}
  
const tabs = document.querySelectorAll('.tab');
tabs.forEach(tab => {
    tab.addEventListener('click', function() {
        const tabName = this.getAttribute('data-name');
        activateTab(tabName);
    });
});

function refreshMembers(data)
{
    if (data.is_owner)
        memberBox.classList.add("auth");
    else 
        memberBox.classList.remove("auth");

    if (memNodes.length > 0) {
        memNodes[0].classList.add("owner");
        memNodes[0].querySelector(".title").textContent = data.mem_usernames[0];
        memNodes[0].querySelector(".after").className = "after fa-solid fa-crown";
    }
    else
        appendNewMember(data.mem_usernames[0], true);

    for (let i = 1; i < data.mem_usernames.length; i++) {
        if (i < memNodes.length) {
            memNodes[i].querySelector(".title").textContent = data.mem_usernames[i];
            memNodes[i].querySelector(".after").className = "after fa-solid fa-times";
        }
        else
            appendNewMember(data.mem_usernames[i], false);

        if (data.is_owner)
            memNodes[i].querySelector(".after").onclick = function() {
                kick(data.mem_ids[i]);
            };
    }
    
    for (let i = data.mem_usernames.length; i < memNodes.length; i++) {
        if (memNodes[i])
            memNodes[i].remove();
        memNodes.splice(i, 1);
    }
}

document.getElementById('input-text').addEventListener('keydown', function(event) {
    if (event.key === 'Enter')
        document.getElementById('sent-message').click();
});

function playMovie(movieId) {
    fetch(webAddr + "/movie/" + movieId)
    .then(res => {
        if (!res.ok)
            alert(`error: ${res.status}`);
        return res.json();
    })
    .then(movie => {
        media.src(movie.url, type='application/x-mpegURL');
        media.currentTime(0);
        media.play();
        updateMovieData();

        mediaContainer.classList.remove("d-none");
        movieGrid.classList.add("d-none");
    })
    .catch(error => alert("Error fetching movie data:" + error));
}

function loadMovies() {
  fetch(webAddr + "/movie")
    .then(res => {
        if (!res.ok)
            alert(`error: ${res.status}`);
        return res.json();
    })
    .then(movies => {
        movies.forEach((movie) => appendNewMovieItem(movie.id, movie.title, movie.poster_url));
    })
    .catch(error => alert("Error fetching movies:" + error));
}

function appendNewMessage(content)
{
    const msgNode = document.createElement("li");
    msgNode.className = "message user-sent";
    msgNode.innerHTML = content;

    const senders = msgBox.querySelectorAll(".sender");
    if (senders.length == 0 || senders[senders.length - 1].innerHTML != "Me") {
        const senderNode = document.createElement("li");
        senderNode.className = "sender user-sent";
        senderNode.innerHTML = "Me";
        msgBox.appendChild(senderNode);
    }

    msgBox.appendChild(msgNode);
}

function appendNewMember(username, isOwner)
{
    const memNode = document.createElement("li");
    memNode.classList.add("member-item");
    if (isOwner)
        memNode.classList.add("owner");
    
    const iconDiv = document.createElement('div');
    iconDiv.className = 'icon';
    
    const icon = document.createElement('i');
    icon.className = 'fa-solid fa-user';
    iconDiv.appendChild(icon);
    
    const titleDiv = document.createElement('div');
    titleDiv.className = 'title';
    titleDiv.textContent = username;

    const after = document.createElement('i');
    if (isOwner)
        after.className = 'after fa-solid fa-crown';
    else
        after.className = 'after fa-solid fa-xmark';
    
    memNode.appendChild(iconDiv);
    memNode.appendChild(titleDiv);
    memNode.appendChild(after);

    memberBox.appendChild(memNode);
    memNodes.push(memNode);
}

function appendNewMovieItem(id, movieTitle, posterUrl) {
    const movieItem = document.createElement('div');
    movieItem.classList.add('movie-item');
    movieItem.setAttribute('data-id', id);

    const movieImage = document.createElement('img');
    movieImage.setAttribute('src', posterUrl);

    const movieTitleElement = document.createElement('h3');
    movieTitleElement.classList.add('movie-title');
    movieTitleElement.textContent = movieTitle;

    movieItem.addEventListener("click", () => {
        const movieId = movieItem.dataset.id;
        if (movieId)
            playMovie(movieId);
        else
            alert("Movie ID not found for this item");
    });

    movieItem.appendChild(movieImage);
    movieItem.appendChild(movieTitleElement);

    movieGrid.appendChild(movieItem);
  }