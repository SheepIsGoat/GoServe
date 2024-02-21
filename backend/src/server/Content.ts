import { httpRequest } from './Handler';

type contentMap<T extends object> = Map<string, T>;
type cache<T extends object> = {
    required: {
        key: keyof T;
        fields: Array<keyof T>;
    };
    store: contentMap<T>; // change this to a connection pool if database
}

type innerFunctionResponse<T> = {
    ok: boolean;
    data?: T;
    errMsg?: string;
}

type contentBusinessLogic<T extends object, U> = (contentCache: cache<T>, data: T) => Promise<U>;
type cblSingleReturn<T extends object> = contentBusinessLogic<T, innerFunctionResponse<T>>;
type cblArrayReturn<T extends object> = contentBusinessLogic<T, innerFunctionResponse<Array<T>>>;


type contentV0 = {
    title: string;
    publishDate: Date;
    owner: string;
    sharingImage1x1Url: string;
    liked: boolean;
    name?: string;
    summary?: string;
    slug?:string;
    status?: string;
}

const requiredContentFields: Array<keyof contentV0> = [
    "title",
    "publishDate",
    "owner",
    "sharingImage1x1Url",
    "liked"
]

function validateContentBody<T extends object>(req: httpRequest, requiredFields: Array<keyof T>): innerFunctionResponse<T> {
    if (!('body' in req)){
        return {
            ok: false,
            errMsg: "No body in request",
        }
    }
    for (const field of requiredFields) {
        if (!(field in req.body)){
            return {
                ok: false,
                errMsg: `"like" request missing required field: ${String(field)}`,
            }
        }
    }
    return {
        data: req.body as T,
        ok: true,
    }
}

export type {
    contentMap, 
    cache, 
    innerFunctionResponse, 
    cblSingleReturn, 
    cblArrayReturn, 
    contentV0
}

export {
    validateContentBody, 
    requiredContentFields
}