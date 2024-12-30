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

const contextMenu = document.getElementById('context-menu');
const addFriendContext = document.getElementById('ctx-add-friend');
const deleteFriendContext = document.getElementById('ctx-delete-friend');
const closeDMContext = document.getElementById('ctx-close-dm');

const sideTabs = document.querySelectorAll(".side-tab");

const lchat = document.getElementById('lchat-box');
const dmContainer = document.querySelector('.msg-ctn');
const chatTabCtn = document.querySelector('.chat-tabs');
const dmInput = document.getElementById('private-sent-input');

const lfriend = document.getElementById('lfriend-box');
const friendList = document.getElementById('friend-list-container');

const tabs = document.querySelectorAll(".tab");

const mediaContainer = document.querySelector(".media-container");
const movieGrid = document.getElementById("movie-grid");
const movieTitle = document.getElementById("movie-title");
const movieSearchbar = document.getElementById('movie-query-input');

const participantNum = document.getElementById("participant-value");
const memberBox = document.getElementById("member-box");
const msgBox = document.getElementById("message-box");
const sentBtn = document.getElementById("sent-message");
const inputText = document.getElementById("input-text");

const webAddr = "http://116.96.94.33:3000";
const webSocketAddr = "ws://116.96.94.33:3000/ws"

const memNodes = []

var activeTab = chatTabCtn.querySelector('.chat-tab.active');
var userId, displayName, boxId, movieId = -1, movieQuery = '', moviePage = 0, movieLoading = false, isDataLoaded = false;
var ws = new WebSocket(webSocketAddr);

ws.onopen = function () {
    const interval = setInterval(() => {
        if (isDataLoaded) {
            clearInterval(interval);
            updateParticipantReq();
            postMovieUpdateRequest();
        }
    }, 400);
    setInterval(() => {
        ws.send(JSON.stringify({ datatype: -1 }));
    }, 10000);
};
ws.onclose = function () {
    checkExists(userId, boxId)
        .then(exists => {
            if (!exists)
                window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
        })
        .catch(error => {
            console.error('Error checking existence:', error);
        });
}

ws.onmessage = function(evt) {
    let data = JSON.parse(evt.data);
    console.log(data);

    if (data.dataType < 0)
        return;

    if (data.datatype == 0)
        processMessageEvent(data);
    else if (data.datatype == 1)
        processMovieEvent(data);
    else if (data.datatype == 2)
        processMemberEvent(data);
    else if (data.datatype == 3)
        processDirectMessageEvent(data);
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
                    userId = data.user_id;
                    displayName = data.display_name;
                    boxId = data.box_id;
                    if (data.is_owner)
                    {
                        ownerLayout(true);
                        setInterval(postMovieData, 2500);
                        media.on("pause", postMovieData);
                        media.on("play", postMovieData);
                    }
                    else
                    {
                        ownerLayout(false);
                    }
                    playMovie(data.movie_id, data.elapsed);
                    loadMessages();
                    document.getElementById("box-id").innerHTML = "#" + boxId;
                    isDataLoaded = true;
                })
                .catch(err => log(err));
            }
        else if (res.status == 401 || res.status == 404)
            window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
        else
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err));
};

function updateParticipantReq() {
    if (ws.readyState != WebSocket.OPEN){
        console.log("ws not ready");
        return;
    }
    ws.send(JSON.stringify(
        {
            datatype: 2
        }
    ));
}

sentBtn.onclick = event => {
    let content = inputText.value;
    if (content == "")
        return;
    
    inputText.value = "";
    appendNewMessage(content, displayName, userId);
    postBoxMessage(content);
}

inputText.addEventListener('keydown', function(event) {
    if (event.key === 'Enter')
        sentBtn.click();
});

dmInput.addEventListener('keydown', function(event) {
    if (event.key === 'Enter') {
        let content = dmInput.value;
        if (content == "")
            return;
        
        dmInput.value = "";
        appendNewDirectMessage(content, true);
        postDirectMessage(activeTab.dataset.userId, content);
    }
});

document.getElementById("expand").onclick = function () {
    media.requestFullscreen();
}

document.getElementById("power-off").onclick = async function () {
    await asyncKick(userId);
}

document.getElementById("leave").onclick = async function () {
    await asyncKick(userId);
}

function kick(id) {
    fetch(`${webAddr}/api/box/${boxId}/user/${id}`, {
        method: 'DELETE',
        credentials: 'include'
    }).then(function(res) {
        if (res.status != 200)
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err))
}

async function asyncKick(id) {
    try {
        const res = await fetch(`${webAddr}/api/box/${boxId}/user/${id}`, {
            method: 'DELETE',
            credentials: 'include'
        });
    
        if (res.status !== 200) {
            const text = await res.text();
            log(text);
        }
    } catch (err) {
        log(err);
    }
}

function processMessageEvent(data) {
    appendNewMessage(data.content, data.display_name, data.sender_id);
}

function processMovieEvent(data) {
    console.log(data);
    media.currentTime(data.elapsed);
    if (data.movie_url == "")
    {
        media.src(data.movie_url);
        mediaContainer.classList.add("d-none");
        movieGrid.classList.remove("d-none");
    }
    else
    {
        mediaContainer.classList.remove("d-none");
        movieGrid.classList.add("d-none");
        movieTitle.innerText = data.movie_title;
        
        if (media.src() != data.movie_url) {
            media.src(data.movie_url);
            media.play();
        }
    }
        
    if (data.is_paused)
        media.pause();
    else 
        media.play();
}

function processMemberEvent(data) {
    participantNum.innerHTML = data.box_user_num;
    checkExists(userId, boxId)
        .then(exists => {
            console.log("receive:");
            console.log(data);
            if (exists)
                loadMembers(data);
            else 
                window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
        })
        .catch(error => {
            console.error('Error checking existence:', error);
        });
}

function processDirectMessageEvent(data) {
    if (data.sender_id == activeTab.dataset.userId)
        appendNewDirectMessage(data.content, data.sender_id == userId);
}

function ownerLayout(isActive) {
    if (isActive) {
        document.getElementById("new-movie").classList.remove("d-none");
        document.getElementById("leave").classList.add("d-none");
        document.getElementById("power-off").classList.remove("d-none");
        movieSearchbar.classList.remove("d-none");
        document.getElementById("new-movie").addEventListener('click', function() {
            movieId = -1;
            media.pause();
            media.src("");
            postMovieData();
            mediaContainer.classList.add("d-none");
            movieGrid.classList.remove("d-none");
            movieSearchbar.classList.remove("d-none");
        });
        media.controls(true);
        loadMovies('', true);
    }
    else {
        document.getElementById("new-movie").classList.add("d-none");
        document.getElementById("leave").classList.remove("d-none");
        document.getElementById("power-off").classList.add("d-none");
        media.controls(false);
        
        movieSearchbar.classList.add("d-none");
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

function activeChatTab(tab) {
    if (activeTab) 
        activeTab.classList.remove('active');
    activeTab = tab;
    if (tab) {
        activeTab.classList.add('active');
        dmInput.classList.remove('d-none');
        loadDirectMessages(activeTab.dataset.userId);
    } else {
        dmInput.classList.add('d-none');
        clearDirectMessages();
    }
}

function activeChatTabById(userId) {
    chatTabCtn.querySelectorAll('.chat-tab').forEach(tab => {
        if (userId == tab.dataset.userId) {
            activeChatTab(tab);
            return;
        }
    });
}

document.addEventListener('click', () => contextMenu.classList.add('d-none'));
addFriendContext.onclick = () => {
    postFriendRequest(contextMenu.dataset.receiverId);
    contextMenu.classList.add('d-none');
    contextMenu.removeAttribute('data-receiver-id');
};
deleteFriendContext.onclick = () => {
    postUnfriend(contextMenu.dataset.friendId);
    closeChatTab(contextMenu.dataset.friendId);
    contextMenu.classList.add('d-none');
    contextMenu.removeAttribute('data-friend-id');
}
closeDMContext.onclick = () => {
    closeChatTab(contextMenu.dataset.friendId);
    contextMenu.classList.add('d-none');
    contextMenu.removeAttribute('data-friend-id');
}

function resetContextMenu(event) {
    event.preventDefault();
    contextMenu.classList.remove('d-none');
    contextMenu.style.left = `${event.pageX}px`;
    contextMenu.style.top = `${event.pageY}px`;

    closeDMContext.classList.add('d-none');
    addFriendContext.classList.add('d-none');
    deleteFriendContext.classList.add('d-none');
}

function setupMemberContext(userID) {
    addFriendContext.classList.remove('d-none');
    contextMenu.dataset.receiverId = userID;
}

function setupFriendContext(userID) {
    deleteFriendContext.classList.remove('d-none');
    contextMenu.dataset.friendId = userID;
}

function setupChatTabContext(userID) {
    closeDMContext.classList.remove('d-none');
    contextMenu.dataset.friendId = userID;
}
  
tabs.forEach(tab => {
    tab.addEventListener('click', function() {
        const tabName = this.getAttribute('data-name');
        activateTab(tabName);
    });
});

document.getElementById('input-text').addEventListener('keydown', function(event) {
    if (event.key === 'Enter')
        document.getElementById('sent-message').click();
});

document.getElementById('friend-toggle').onclick = () => sideTabToggle(lfriend);
document.getElementById('chat-toggle').onclick = () => sideTabToggle(lchat);
document.querySelectorAll('.chat-tab').forEach(tab => tab.onclick = () => activeChatTab(tab));

movieSearchbar.onchange = event => {
    if (movieQuery != event.target.value) {
        loadMovies(event.target.value, true);
        movieQuery = event.target.value;
    }
};

movieGrid.addEventListener('scroll', () => {
    if (!movieLoading && movieGrid.scrollHeight - movieGrid.scrollTop === movieGrid.clientHeight)
        loadMovies(movieQuery, false);
});

function sideTabToggle(sideTab) {
    if (sideTab == lchat) {
        if (!activeTab)
            clearDirectMessages();
    } else if (sideTab == lfriend)
        loadFriends();

    sideTabs.forEach(tab => {
        if (tab != sideTab)
            tab.classList.remove("open");
    });
    sideTab.classList.toggle('open');
}

function playMovie(movieID, elapsed) {
    if (movieID < 0) {
        movieId = movieID;
        return;
    }
    fetch(webAddr + "/api/movie/" + movieID)
    .then(res => {
        if (!res.ok)
            log(`error: ${res.status}`);
        return res.json();
    })
    .then(movie => {
        movieId = movieID;
        media.src(movie.url, type='application/x-mpegURL');
        media.currentTime(elapsed);
        media.play();
        postMovieData();

        movieTitle.innerText = movie.title;
        mediaContainer.classList.remove("d-none");
        movieGrid.classList.add("d-none");
        movieSearchbar.classList.add("d-none");
    })
    .catch(error => log("Error fetching movie data:" + error));
}

async function checkExists(userId, boxId) {
    try {
        const res = await fetch(`${webAddr}/api/box/${boxId}/exists/${userId}`, {
            method: 'GET',
            credentials: 'include',
            headers: {
                'Cache-Control': 'must-revalidate,no-cache,no-store',
                'Pragma': 'no-cache',
            }
        });

        if (res.status !== 200) {
            const err = await res.text();
            log(err);
            throw new Error(`HTTP Error: ${res.status}`);
        }

        const json = await res.json();
        return json.value;
    } catch (error) {
        log(error);
        return false;
    }
}

function loadMovies(query, clear) {
    movieLoading = true;
    const pageTemp = clear ? 0 : moviePage;
    fetch(webAddr + "/api/movie/search?page=" + pageTemp + "&query=" + query)
    .then(res => {
        if (!res.ok)
            log(`error: ${res.status}`);
        return res.json();
    })
    .then(movies => {
        if (clear) {
            moviePage = 1;
            clearMovies();
        } 
        else
            moviePage++;
        movies.forEach(movie => appendNewMovieItem(movie.id, movie.title, movie.poster_url));
        movieLoading = false;
    })
    .catch(error => {
        log("Error fetching movies:" + error);
        movieLoading = false;
    });
}

function loadMessages() {
    fetch(`${webAddr}/api/box/${boxId}/msg`)
    .then(res => {
        if (!res.ok)
            log(`error: ${res.status}`);
        return res.json();
    })
    .then(messages => {
        messages.forEach(msg => appendNewMessage(msg.content, msg.display_name, msg.user_id));
    })
    .catch(error => log("Error fetching messages in box:" + error));
}

function loadDirectMessages(targetId) {
    fetch(`${webAddr}/api/user/${userId}/msg/${targetId}`)
    .then(res => {
        if (!res.ok)
            log(`error: ${res.status}`);
        return res.json();
    })
    .then(messages => {
        clearDirectMessages();
        messages.forEach(msg => appendNewDirectMessage(msg.content, userId == msg.user_id));
    })
    .catch(error => log("Error fetching direct messages with user #" + targetId + ": " + error));
}

function loadFriends() {
    fetch(`${webAddr}/api/user/${userId}/friends`)
    .then(res => {
        if (!res.ok)
            log(`error: ${res.status}`);
        return res.json();
    })
    .then(friends => {
        console.log(friends);
        clearFriendList();
        friends.forEach(friend => appendNewFriend(friend.display_name, friend.is_online, friend.user_id));
    })
    .catch(error => log("Error fetching friends of user #" + userId + ": " + error));
}

function loadMembers(data) {
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
        appendNewMember(data.mem_usernames[0], data.mem_ids[0], true);

    for (let i = 1; i < data.mem_usernames.length; i++) {
        if (i < memNodes.length) {
            memNodes[i].querySelector(".title").textContent = data.mem_usernames[i];
            memNodes[i].querySelector(".after").className = "after fa-solid fa-times";
        }
        else
            appendNewMember(data.mem_usernames[i], data.mem_ids[i], false);

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

function closeChatTab(friendId) {
    chatTabCtn.querySelectorAll(".chat-tab").forEach(tab => {
        if (tab.dataset.userId == friendId) {
            tab.remove();
            if (tab == activateTab)
                activeChatTab(undefined);
            return;
        }
    });
}

function postMovieUpdateRequest() {
    fetch(`${webAddr}/api/box/${boxId}/movie/update`, {
        method: 'POST',
        credentials: 'include',
    }).then(function(res) {
        if (res.status != 200)
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err))
}

function postUnfriend(friendId) {
    fetch(`${webAddr}/api/user/${userId}/friends/${friendId}`, {
        method: 'DELETE',
        credentials: 'include',
    }).then(function(res) {
        if (res.status == 200)
            loadFriends();
        else 
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err))
}

function postFriendRequest(receiverId) {
    fetch(`${webAddr}/api/user/${userId}/add/${receiverId}`, {
        method: 'POST',
        credentials: 'include',
    }).then(function(res) {
        if (res.status == 200 || res.status == 201)
            loadFriends();
        else
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err))
}

function postMovieData() {
    fetch(`${webAddr}/api/box/${boxId}/movie`, {
        method: 'PATCH',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            movie_id: parseInt(movieId),
            elapsed: media.currentTime(),
            is_paused: media.paused()
        })
    }).then(function(res) {
        if (res.status == 401)
            window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
        else if (res.status != 200)
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err))
}

function postBoxMessage(content) {
    console.log(JSON.stringify({
        user_id: userId,
        box_id: boxId,
        content: content
    }));
    fetch(`${webAddr}/api/box/${boxId}/msg/${userId}`, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            content: content
        })
    }).then(function(res) {
        if (res.status != 201)
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err))
}

function postDirectMessage(receiverId, content) {
    console.log(JSON.stringify({
        user_id: userId,
        receiver_id: receiverId,
        content: content
    }));
    fetch(`${webAddr}/api/user/${userId}/msg/${receiverId}`, {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            content: content
        })
    }).then(function(res) {
        if (res.status != 201)
            res.text()
                .then(text => log(text))
                .catch(err => log(err));
    })
    .catch(err => log(err))
}

function clearMovies() {
    movieGrid.innerHTML = '';
}

function clearDirectMessages() {
    dmContainer.innerHTML = '';
}

function clearFriendList() {
    friendList.innerHTML = '';
}

function appendNewChatTab(displayName, userId) {
    const newTab = document.createElement('div');
    newTab.classList.add('chat-tab');
    newTab.dataset.userId = userId;
    newTab.textContent = displayName;
    
    newTab.addEventListener('contextmenu', function (event) {
        resetContextMenu(event);
        setupChatTabContext(userId);
    });

    chatTabCtn.appendChild(newTab);
}

function appendNewMessage(content, displayName, senderId) {
    const msgNode = document.createElement("li");
    msgNode.innerHTML = content;

    if (userId == senderId) {
        msgNode.className = "message user-sent";
        displayName = "Me";
    } else {
        msgNode.className = "message";
    }

    const senders = msgBox.querySelectorAll(".sender");
    if (senders.length == 0 || senders[senders.length - 1].dataset.senderId != senderId) {
        const senderNode = document.createElement("li");
        senderNode.dataset.senderId = senderId;
        senderNode.className = userId == senderId ? "sender user-sent" : "sender";
        senderNode.innerHTML = displayName;
        msgBox.appendChild(senderNode);
    }

    msgBox.appendChild(msgNode);
}

function appendNewMember(displayName, userId, isOwner) {
    const memNode = document.createElement("li");
    memNode.dataset.userId = userId;
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
    titleDiv.textContent = displayName;

    const after = document.createElement('i');
    if (isOwner)
        after.className = 'after fa-solid fa-crown';
    else
        after.className = 'after fa-solid fa-xmark';

    memNode.appendChild(iconDiv);
    memNode.appendChild(titleDiv);
    memNode.appendChild(after);
    
    memNode.addEventListener('contextmenu', function (event) {
        resetContextMenu(event);
        setupMemberContext(userId);
    });
    
    memberBox.appendChild(memNode);
    memNodes.push(memNode);
}

function appendNewMovieItem(id, movieTitle, posterUrl) {
    const movieItem = document.createElement('div');
    movieItem.classList.add('movie-item');
    movieItem.dataset.id = id;

    const movieImage = document.createElement('img');
    movieImage.setAttribute('src', posterUrl);

    const movieTitleElement = document.createElement('h3');
    movieTitleElement.classList.add('movie-title');
    movieTitleElement.textContent = movieTitle;

    movieItem.addEventListener("click", () => {
        const mid = movieItem.dataset.id;
        if (mid)
            playMovie(mid, 0);
        else
            log("Movie ID not found for this item");
    });

    movieItem.appendChild(movieImage);
    movieItem.appendChild(movieTitleElement);

    movieGrid.appendChild(movieItem);
}

function appendNewFriend(friendName, isOnline, friendId) {
    const friendItem = document.createElement('li');
    friendItem.className = `friend-item ${isOnline ? 'online' : ''}`;
    
    const nameSpan = document.createElement('span');
    nameSpan.className = 'friend-name';
    nameSpan.textContent = friendName + ' #' + friendId;
    
    const statusSpan = document.createElement('span');
    statusSpan.className = 'friend-status online-status-dot';
    
    const dmButton = document.createElement('button');
    dmButton.className = 'dm-icon-button';
    dmButton.dataset.userId = friendId;
    
    const dmIcon = document.createElement('i');
    dmIcon.className = 'fa-solid fa-comment icon';
    
    dmButton.appendChild(dmIcon);
    
    dmButton.onclick = () => {
        let found;
        chatTabCtn.querySelectorAll(".chat-tab").forEach(tab => {
            if (tab.dataset.userId == friendId) {
                found = tab;
                return;
            }
        });
        if (found)
            activeChatTab(found);
        else {
            appendNewChatTab(friendName, friendId);
            activeChatTabById(friendId);
        }
        sideTabToggle(lchat);
    };

    friendItem.appendChild(nameSpan);
    friendItem.appendChild(statusSpan);
    friendItem.appendChild(dmButton);
    
    friendItem.addEventListener('contextmenu', function (event) {
        resetContextMenu(event);
        setupFriendContext(friendId);
    });

    friendList.appendChild(friendItem);
}

function appendNewDirectMessage(content, isUser = false) {
    const messageItem = document.createElement('li');
    messageItem.className = `message${isUser ? ' user-sent' : ''}`;
    messageItem.textContent = content;

    dmContainer.appendChild(messageItem);
}

function log(msg) {
    console.log(msg);
    alert(msg);
}