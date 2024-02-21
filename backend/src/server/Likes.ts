import { 
    httpResult, 
    httpRequest, 
    requestHandler 
} from './Handler';

import {
    contentMap, 
    cache, 
    innerFunctionResponse, 
    cblSingleReturn,
    cblArrayReturn,
    validateContentBody, 
    contentV0, 
    requiredContentFields
}  from './Content';


// state

const likeCacheMap: contentMap<contentV0> = new Map<string, contentV0>() ;  // currently a global cache, change to a db in prod

const GlobalLikeCache: cache<contentV0> = {
    required: {
        key: "title",
        fields: requiredContentFields,
    },
    store: likeCacheMap
}

// functions

// entrypoint functions
const getLiked: requestHandler<{}> = async ({}) => {
    return getLikedHandler<contentV0>(GlobalLikeCache);
}

const like: requestHandler<httpRequest> = async (request: httpRequest) => {
     return unLikeHandler<contentV0>(request, likeFunc, GlobalLikeCache)
}

const unlike: requestHandler<httpRequest> = async (request: httpRequest) => {
    return unLikeHandler<contentV0>(request, unlikeFunc, GlobalLikeCache)
}

// helper functions
async function unLikeHandler<T extends object>(request: httpRequest, likeAction: cblSingleReturn<T>, likeCache: cache<T>): Promise<httpResult> {
    // for '/like' and '/unlike'. Easily extensible to handle all content interaction requests - i.e. views, comments, etc.
    // T should be an object that describes the API json schema, like contentV0
    const bodyValidation: innerFunctionResponse<T> = validateContentBody<T>(request, likeCache.required.fields);
    if (!bodyValidation.ok) {
        return {
            data: {},
            statusCode: 400,
            meta: {
                errMsg: `Error: ${bodyValidation.errMsg}. Function: ${likeAction.name}`
            }
        }
    }
    const likeValidation: innerFunctionResponse<T> = await likeAction(likeCache, bodyValidation.data as T);
    if (!likeValidation.ok){
        return {
            data: {},
            statusCode: 500,
            meta: {
                errMsg: `Internal server error. Operation failed with function: ${likeAction.name}.`
            }
        }
    }
    return {
        data: {},
        statusCode: 200,
    }
}

const likeFunc: cblSingleReturn<any> = async<T extends object> (likeCache: cache<T>, data: T) => {
    // likes the post in the cache
    const key = String(data[likeCache.required.key as keyof T]);
    likeCache.store.set(key, data);
    return {ok: true};
}

const unlikeFunc: cblSingleReturn<any> = async<T extends object> (likeCache: cache<T>, data: T) => {
    // unlikes the post in the cache
    const key = String(data[likeCache.required.key as keyof T]);
    likeCache.store.delete(key);
    return {ok: true}
}

async function getLikedHandler<T extends object>(likeCache: cache<T>): Promise<httpResult>{
    const validation: innerFunctionResponse<Array<T>> = await getLikedFunc(likeCache, null);
    if (!validation.ok){
        return {
            statusCode: 500,
            meta: {
                errMsg: `Internal server error. Operation failed with function: ${getLikedFunc.name}.`
            }
        }
    }
    return {
        data: validation.data,
        statusCode: 200,
    }
}

const getLikedFunc: cblArrayReturn<any> = async<T extends object> (likeCache: cache<T>) => {
    // returns the whole cache as an array
    return {
        ok: true,
        data: Array.from(likeCache.store.values())
    }
}

export {likeCacheMap as likeCache, getLiked, like, unlike};