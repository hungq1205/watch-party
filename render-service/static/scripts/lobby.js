let isCreate = false;
const title = document.getElementById("operation-title");
const alterOption = document.getElementById("alter-option");
const boxId = document.getElementById("box-id");
const submit = document.getElementById("submit");
const form = document.querySelector("form");

const webAddr = "116.96.94.33";

window.onload = function() {
    alterOption.onclick = function() {
        isCreate = !isCreate;
        createForm(isCreate);
    };
};

form.addEventListener("submit", e => {
    e.preventDefault();
    const formData = new FormData(form);
    const data = new URLSearchParams(formData);

    data.set("box_id", data.get("box_id").trim())

    if (isCreate) {
        data.delete("box_id");
        fetch('http://' + webAddr + ':3000/create', {
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
                window.location.replace('http://' + webAddr + ':3000/login?nocache=' + new Date().getTime());
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
    fetch('http://' + webAddr + ':3000/join', {
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