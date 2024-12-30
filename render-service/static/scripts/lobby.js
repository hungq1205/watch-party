const pwPopup = document.getElementById("password-popup");
const pwLabel = document.getElementById("password-label");
const pwInput = document.getElementById("password-input");
const overlay = document.getElementById("overlay");

const boxCtn = document.getElementById('box-container');

const contextMenu = document.getElementById('context-menu');
const deleteFriendContext = document.getElementById('ctx-delete-friend');
const closeDMContext = document.getElementById('ctx-close-dm');

const sideTabs = document.querySelectorAll(".side-tab");

const lchat = document.getElementById('lchat-box');
const dmContainer = document.querySelector('.msg-ctn');
const chatTabCtn = document.querySelector('.chat-tabs');
const dmInput = document.getElementById('private-sent-input');

const lfriend = document.getElementById('lfriend-box');
const friendList = document.getElementById('friend-list-container');

const webAddr = "http://116.96.94.33:3000";

var activeTab = chatTabCtn.querySelector('.chat-tab.active');
var ws = new WebSocket("ws://116.96.94.33:3000/ws");
var userId, isDataLoaded = false, boxJoinId = -1;

ws.onopen = function () {
    setInterval(() => {
        ws.send(JSON.stringify({ datatype: -1 }));
    }, 10000);
};
ws.onclose = function () {
    window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
}

ws.onmessage = function(evt) {
    let data = JSON.parse(evt.data);
    console.log(data);

    if (data.dataType < 0)
        return;

    if (data.datatype == 3)
        processDirectMessageEvent(data);
};

window.onload = function() {
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
                isDataLoaded = true;
                loadBoxes();
                setInterval(loadBoxes, 4000);
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

pwPopup.addEventListener("submit", e => {
    e.preventDefault();
    const formData = new FormData(pwPopup);
    const data = new URLSearchParams(formData);

    if (boxJoinId < 0) {
        fetch(webAddr + '/api/box', {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'Cache-Control': 'must-revalidate,no-cache,no-store'
            },
            body: data
        }).then(function(res) {
            if (res.status == 201 || res.status == 200) 
            {
                res.json().then(function(jdata) {
                    data.append("box_id", jdata["id"]);
                    joinBox(data);
                });
            }
            else if (res.status == 401)
            {
                window.location.replace(webAddr + '/login?nocache=' + new Date().getTime());
            }
            else
                res.text()
                    .then(text => alert(text))
                    .catch(err => alert(err));
        })
        .catch(err => alert(err));
    }
    else {
        data.append("box_id", boxJoinId);
        joinBox(data);
    }
    closePwPopup();
});

function joinBox(fdata) {
    fetch(webAddr + '/join', {
        method: 'POST',
        credentials: 'include',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'Cache-Control': 'must-revalidate,no-cache,no-store'
        },
        body: fdata
    }).then(function(res) {
        if (res.status == 200) {
            window.location.reload();
        }
        else {
            window.location.reload();
            res.text()
            .then(text => alert(text))
            .catch(err => alert(err));
        }
    })
    .catch(err => alert(err));
}

function processDirectMessageEvent(data) {
    if (data.sender_id == activeTab.dataset.userId)
        appendNewDirectMessage(data.content, data.sender_id == userId);
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

function sideTabToggle(sideTab) {
    if (sideTab == lchat) {
        if (!activeTab)
            clearDirectMessages();
    } else if (sideTab == lfriend)
        loadFriends();

    sideTabs.forEach(tab => {
        if (tab != sideTab)
            tab.classList.remove('open');
    });
    sideTab.classList.toggle('open');
}

function openPwPopup(label) {
    overlay.style.display = 'block';
    pwPopup.style.display = 'block';
    pwLabel.innerText = label;
    pwInput.focus();
}

function closePwPopup() {
    overlay.style.display = 'none';
    pwPopup.style.display = 'none';
    boxJoinId = -1;
}

document.addEventListener('click', () => contextMenu.classList.add('d-none'));
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
    deleteFriendContext.classList.add('d-none');
}

function setupFriendContext(userID) {
    deleteFriendContext.classList.remove('d-none');
    contextMenu.dataset.friendId = userID;
}

function setupChatTabContext(userID) {
    closeDMContext.classList.remove('d-none');
    contextMenu.dataset.friendId = userID;
}

document.getElementById('friend-toggle').onclick = () => sideTabToggle(lfriend);
document.getElementById('chat-toggle').onclick = () => sideTabToggle(lchat);
document.querySelectorAll('.chat-tab').forEach(tab => tab.onclick = () => activeChatTab(tab));

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

function loadBoxes() {
    fetch(`${webAddr}/api/box`)
    .then(res => {
        if (!res.ok)
            log(`error: ${res.status}`);
        return res.json();
    })
    .then(boxes => {
        console.log(boxes);
        clearBoxList();
        boxes.forEach(box => appendNewBoxItem(box));
    })
    .catch(error => log("Error fetching boxes: " + error));
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

function clearDirectMessages() {
    dmContainer.innerHTML = '';
}

function clearFriendList() {
    friendList.innerHTML = '';
}

function clearBoxList() {
    boxCtn.innerHTML = '';
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

function appendNewBoxItem(data) {
    const movieBox = document.createElement('div');
    movieBox.className = 'movie-box';
    movieBox.dataset.id = data.box_id;

    const img = document.createElement('img');
    img.src = data.movie_poster_url == "" ? "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSzMP_KMX-JYjb9tSoCTdzSNlC9BKI9rSBM7Q&s" : data.movie_poster_url;
    img.className = 'poster';

    const infoDiv = document.createElement('div');
    infoDiv.className = 'info';

    const title = document.createElement('h3');
    title.className = 'title';
    title.textContent = data.movie_title;

    const owner = document.createElement('p');
    owner.className = 'username';
    owner.textContent = `Owner: ${data.owner_display_name}`;

    const members = document.createElement('p');
    members.className = 'members';
    members.innerHTML = `<i class="fas fa-users"></i> ${data.member_num}`;

    const progressBar = document.createElement('div');
    progressBar.className = 'progress-bar';

    const progress = document.createElement('div');
    progress.className = 'progress';
    progress.style.width = `${data.elapsed / 72.0}%`;

    progressBar.appendChild(progress);

    infoDiv.appendChild(title);
    infoDiv.appendChild(owner);
    infoDiv.appendChild(members);
    infoDiv.appendChild(progressBar);

    movieBox.appendChild(img);
    movieBox.appendChild(infoDiv);

    movieBox.onclick = event => {
        boxJoinId = movieBox.dataset.id;
        openPwPopup("Enter password");
    };

    boxCtn.appendChild(movieBox);
}

function log(msg) {
    console.log(msg);
    alert(msg);
}