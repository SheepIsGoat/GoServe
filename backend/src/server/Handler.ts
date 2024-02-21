import { Request, Response } from 'express';

type requestHandler<T> = (args: T) => Promise<httpResult>;
interface httpResult {
    statusCode: httpStatusCode;
    data?: any;
    meta?: {
        errMsg?: string;
    };
}

type httpRequest = Request;

type httpStatusCode = 200 | 400 | 500;

async function handle(funcs: Array<requestHandler<httpRequest>>, req: httpRequest, res: Response){
    // handle an array of functions and return a Response with status code and data
    try {
        let result: httpResult = {
            data: {},
            statusCode: 500,
        }
        for (const func of funcs) {
            result = await func(req);
            if (result.statusCode >= 400) {
                return res.status(result.statusCode).json(result.meta?.errMsg);
            }
        }
        return res.status(result.statusCode).json(result.data);
    } catch(error) {
        console.error('Error:', error);
        res.status(500).json({ error: 'Internal Server Error' });
    }
}

export type {requestHandler, httpResult, httpRequest, httpStatusCode};
export {handle};

