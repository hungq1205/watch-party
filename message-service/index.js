const PORT = 3002;
const PREFIX = "/api/msg"

const { Pool } = require('pg');

const pool = new Pool({
    user: 'postgres',
    host: 'localhost',
    database: 'message_service',
    password: 'hungthoi',
    port: 5432,
});

pool
    .connect()
    .then(client => {
        console.log('Connected to PostgreSQL');
        client.release();
    })
    .catch(err => console.error('Connection error', err.stack));

const express = require('express');
const bodyParser = require('body-parser');
  
const app = express();

app.use(bodyParser.json());

// filter messages
app.get(PREFIX + '', async (req, res) => {
    const { box_id, user_id } = req.query;
    try {
        let msgs;
        if (box_id == undefined)
        {
            if (user_id == undefined)
                res.status(400).json({ message: 'Bad request' });
            else
                msgs = await pool.query('SELECT * FROM message WHERE user_id = $1', [user_id]);
        }
        else if (user_id == undefined)
            msgs = await pool.query(
                `SELECT m.*
                FROM message m
                JOIN box_msg bm ON bm.msg_id = m.id
                WHERE bm.box_id = $1`,
                [box_id]);
        else
            msgs = await pool.query(
                `SELECT m.*
                FROM message m
                JOIN box_msg bm ON bm.msg_id = m.id
                WHERE bm.box_id = $1 AND m.user_id = $2`,
                [box_id, user_id]);
        res.status(200).json(msgs.rows);
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// get direct messages of 2 users
app.get(PREFIX + '/dm', async (req, res) => {
    const { user_id1, user_id2 } = req.query;
    try {
        if (user_id1 == undefined || user_id2 == undefined)
            res.status(400);
        else {
            boxId = await pool.query(
                `SELECT m.*
                FROM message m
                JOIN box_msg bm ON bm.msg_id = m.id
                JOIN box b ON b.id = bm.box_id
                JOIN box_user bu1 ON bu1.box_id = b.id
                JOIN box_user bu2 ON bu2.box_id = b.id
                WHERE b.isDM and bu1.user_id = $1 AND bu2.user_id = $2`,
                [user_id1, user_id2]);
            res.status(200).json(boxId.rows);
        }
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// get message by id
app.get(PREFIX + '/:id', async (req, res) => {
    const { id } = req.params;
    try {
        const result = await pool.query('SELECT * FROM message WHERE id = $1', [id]);
        if (result.rows.length > 0) {
            res.json(result.rows[0]);
        } else {
            res.status(404).json({ message: 'Message not found' });
        }
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// create message
app.post(PREFIX, async (req, res) => {
    const { box_id, user_id, content, sender_id, receiver_id } = req.body;
    try {
        if (user_id == undefined) {
            await pool.query(
                'call sent_dm($1, $2, $3)',
                [sender_id, receiver_id, content]
            );
            res.status(201).json({ message: 'Message created' });
        }
        else {
            await pool.query(
                'call sent_msg_to_box($1, $2, $3)',
                [user_id, box_id, content]
            );
            res.status(201).json({ message: 'Message created' });
        }
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// get users in box
app.get(PREFIX + '/box/:id', async (req, res) => {
    const { id } = req.params;
    try {
        const result = await pool.query(
            `SELECT bu.user_id
            FROM box b
            JOIN box_user bu ON b.id = bu.box_id  
            WHERE b.id = $1`, 
            [id]);
        res.status(200).json({ box_id: id, user_ids: result.rows });
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// delete message
app.delete(PREFIX + '/:id', async (req, res) => {
    const { id } = req.params;
    try {
        const result = await pool.query('DELETE FROM message WHERE id = $1 RETURNING *', [id]);
        await pool.query('DELETE FROM box_msg WHERE msg_id = $1', [id]);
        if (result.rows.length > 0) {
            res.json({ id: result.rows[0] });
        } else {
            res.status(404).json({ message: 'Messge not found' });
        }
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// create box
app.post(PREFIX + "/box", async (req, res) => {
    const { user_id } = req.body;
    try {
        const box = await pool.query('INSERT INTO box DEFAULT VALUES RETURNING *');
        const box_id = box.rows[0].id;
        res.status(201).json({ id: box_id });
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// delete box
app.delete(PREFIX + "/box/:id", async (req, res) => {
    const { id } = req.params;
    try {
        await pool.query(
            `DELETE FROM message
            USING box_msg bm
            WHERE bm.msg_id = message.id
            AND bm.box_id = $1;`
            , [id]);
        await pool.query('DELETE FROM box WHERE id = $1', [id]);
        await pool.query('DELETE FROM box_user WHERE box_id = $1', [id]);
        await pool.query('DELETE FROM box_msg WHERE box_id = $1', [id]);
        res.status(200).json({ message: 'Box ' + id + ' deleted successfully' });
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// add user to box
app.post(PREFIX + "/box/:box_id/add/:user_id", async (req, res) => {
    const { box_id, user_id } = req.params;
    try {
        await pool.query(
            'INSERT INTO box_user (box_id, user_id) VALUES ($1, $2)',
            [box_id, user_id]
        );
        res.status(200).json({ message: 'User' + user_id + ' added to box ' + box_id + ' successfully'});
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

// remove user from box
app.delete(PREFIX + "/box/:box_id/remove/:user_id", async (req, res) => {
    const { box_id, user_id } = req.params;
    try {
        await pool.query(
            'DELETE FROM box_user WHERE box_id = $1 AND user_id = $2',
            [box_id, user_id]
        );
        res.status(200).json({ message: 'User' + user_id + ' removed from box ' + box_id + ' successfully'});
    } catch (err) {
        console.log(err)
        res.status(500).json({ error: err.message });
    }
});

app.listen(PORT, () => {
    console.log(`Message service started on http://localhost:${PORT}`);
});