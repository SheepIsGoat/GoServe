const dotenv = require('dotenv');
dotenv.config({ path: './process.env' });

const express = require('express');
const path = require('path');

import { Request, Response, NextFunction } from 'express';

import { handle } from './server/Handler';
import { searchSanity } from './server/Sanity';
import { likeCache, getLiked, like, unlike } from './server/Likes';
// global mutable state likeCache, declared for dependency tracking. It is initialized in its own file.
likeCache;

const app = express();
const PORT: number = 3000;

app.use(express.json());
app.use(express.static(path.join(__dirname, 'build')));
app.use((_: null, res: Response, next: NextFunction) => {
    res.header('Access-Control-Allow-Origin', '*'); // Allow requests from any origin
    res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS'); // Allow specified HTTP methods
    res.header('Access-Control-Allow-Headers', 'Origin, X-Requested-With, Content-Type, Accept'); // Allow specified headers
    next();
});


app.listen(PORT, () => {
  console.log(`Server is running on http://localhost:${PORT}`);
});

app.get('/search', async (req: Request, res: Response) => {
    handle([searchSanity], req, res);
});

app.post('/like', async (req: Request, res: Response) => {
    handle([like, getLiked], req, res);
});

app.post('/unlike', async (req: Request, res: Response) => {
    handle([unlike, getLiked], req, res);
});

app.get('/get_liked', async (req: Request, res: Response) => {
    handle([getLiked], req, res);
});

export {};