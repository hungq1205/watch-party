let isCreate = false;
const title = document.getElementById("operation-title");
const alterOption = document.getElementById("alter-option");
const boxId = document.getElementById("box-id");
const submit = document.getElementById("submit");
const form = document.querySelector("form");

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
var userId, isDataLoaded = false;

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
    alterOption.onclick = function() {
        isCreate = !isCreate;
        createForm(isCreate);
    };
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

form.addEventListener("submit", e => {
    e.preventDefault();
    const formData = new FormData(form);
    const data = new URLSearchParams(formData);

    data.set("box_id", data.get("box_id").trim())

    if (isCreate) {
        data.delete("box_id");
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
                alert("Created");
                res.json().then(function(jdata) {
                    data.append("box_id", jdata["id"])
                    joinBox(data)
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
        .catch(err => alert(err))
    }
    else {
        joinBox(data)
    }
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

function createForm(isActive)
{
    if (isActive)
    {
        title.textContent = "Create Box"
        alterOption.textContent = "Join";
        boxId.parentElement.classList.add("d-none");
        submit.textContent = "Create";
    }
    else
    {
        title.textContent = "Join Box"
        alterOption.textContent = "Create";
        boxId.parentElement.classList.remove("d-none");
        submit.textContent = "Join";
    }
}

function activeChatTab(tab) {
    if (activeTab) 
        activeTab.classList.remove('active');
    activeTab = tab;
    if (activeTab) {
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

function log(msg) {
    console.log(msg);
    alert(msg);
}