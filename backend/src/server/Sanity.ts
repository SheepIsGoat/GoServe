import axios, { AxiosResponse } from 'axios';
import { httpRequest, httpResult, requestHandler } from './Handler';
import {
    contentMap, 
    cache, 
    innerFunctionResponse, 
}  from './Content';

interface response {
    data: any,
    status: number
}

const SANITY_TOKEN: string = String(process.env.SANITY_TOKEN);

const searchSanity: requestHandler<httpRequest> = async (request: httpRequest) => {
    return searchHandler<AxiosResponse>(request);
}

async function searchHandler<T extends response>(request: httpRequest): Promise<httpResult>{
    // T is for Sanity SDK
    const qValidation = validateQueryString<string>(request);
    if (!qValidation.ok) {
        return {
            statusCode: 400,
            data: {},
            meta: {
                errMsg: `Error: ${qValidation.errMsg}. Function: ${searchHandler}`
            }
        }
    }
    const searchTerm = String(qValidation.data);
    
    const sanityQuery = buildSanityQuery(searchTerm);
    const sanityResponse = await querySanity<T>(sanityQuery);
    const processedSanityResponse = processSanityResponse<T>(sanityResponse);
    console.log("Returning query results: " + processedSanityResponse)
    return {
        statusCode: 200,
        data: processedSanityResponse,
    }
}

function validateQueryString<T>(request: httpRequest): innerFunctionResponse<T>{
    if (!('query' in request)){
        return {
            ok: false,
            errMsg: "No search query in request",
        }
    }
    if (!('q' in request.query)){
        return {
            ok: false,
            errMsg: "No .q property in in search request.query",
        }
    }
    return {
        ok: true,
        data: request.query.q as T,
    }
}

function buildSanityQuery(searchTerm: string): string {
    console.log("Received search request " + searchTerm);
    const cleanSearchTerm: string = cleanQuery(searchTerm);
    const query: string = `*[_type == "tile" && title match "${cleanSearchTerm}*"]{name,title,summary,slug,status,publishDate,owner,sharingImage1x1Url}`;
    return encodeURIComponent(query);
}

async function querySanity<T>(sanityQuery: string): Promise<T> {
    // This function is not actually generic, it uses axios.get no matter what.
    // TODO: build url from env
    const url = `https://0unlbb72.api.sanity.io/v2022-03-29/data/query/dev?query=${sanityQuery}`;
    const options = {
        headers: {
            'Authorization': `Bearer ${SANITY_TOKEN}`
        }
    }
    return await axios.get(url, options);
    
}

function processSanityResponse<T extends response>(response: T): innerFunctionResponse<object> {
    let data = response.data;
    delete data.query;
    return {
        ok: response.status < 400,
        data: data,
        errMsg: response.status >= 400? `Status code: ${response.status}, message: ${data}`: undefined,
    }
}

const naughtyChars = new Set(['\0', '\x08', '\x09', '\x1a', '\n', '\r', '"', "'", '\\', '%', '[', ']', '{', '}', ',']);
function cleanQuery(queryString: string): string {
    // Should use official list in prod, this is just a small example
    // https://github.com/minimaxir/big-list-of-naughty-strings
    
    let sanitizedQuery = '';
    for (const char of queryString) {
        if (!naughtyChars.has(char)) {
            sanitizedQuery += char;
        }
    }

    return sanitizedQuery;
}

export {searchSanity};