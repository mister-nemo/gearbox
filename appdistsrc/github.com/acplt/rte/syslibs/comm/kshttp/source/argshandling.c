/*
*	Copyright (C) 2015
*	Chair of Process Control Engineering,
*	Aachen University of Technology.
*	All rights reserved.
*
*	Redistribution and use in source and binary forms, with or without
*	modification, are permitted provided that the following conditions
*	are met:
*	1. Redistributions of source code must retain the above copyright
*	   notice, this list of conditions and the following disclaimer.
*	2. Redistributions in binary form must print or display the above
*	   copyright notice either during startup or must have a means for
*	   the user to view the copyright notice.
*	3. Redistributions in binary form must reproduce the above copyright
*	   notice, this list of conditions and the following disclaimer in
*		the documentation and/or other materials provided with the
*		distribution.
*	4. Neither the name of the Chair of Process Control Engineering nor
*		the name of the Aachen University of Technology may be used to
*		endorse or promote products derived from this software without
*		specific prior written permission.
*
*	THIS SOFTWARE IS PROVIDED BY THE CHAIR OF PROCESS CONTROL ENGINEERING
*	``AS IS'' AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
*	LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
*	A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE CHAIR OF
*	PROCESS CONTROL ENGINEERING BE LIABLE FOR ANY DIRECT, INDIRECT,
*	INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
*	BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
*	OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED
*	AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
*	LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY
*	WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
*	POSSIBILITY OF SUCH DAMAGE.
*
***********************************************************************/

#include "config.h"
#include "urldecode.h"
#include <errno.h>

/**
 * This function will searches for matching varname(s) and returns a list of values
 * Note: varnames like var[0], var[1], var[xx], var[ are matched and returned in the list
 *
 * @param urlQuery argument pair array returned by parse_http_request
 * @param varname mask to search like "varname"
 * @param re return list
 * @return always success
 */
OV_RESULT kshttp_find_arguments(const OV_STRING_VEC* urlQuery, const OV_STRING varname, OV_STRING_VEC* re){
	OV_UINT i = 0;
	OV_UINT varname_len = 0;
	Ov_SetDynamicVectorLength(re,0,STRING);	//initialize the return vector properly
	if(urlQuery == NULL || varname == NULL){
		return OV_ERR_OK;
	}
	//iterate over argument names
	for(i = 0;i < urlQuery->veclen;i=i+2){
		//simple match -> put value in the output list
		varname_len = ov_string_getlength(varname);
		if(strncmp(urlQuery->value[i], varname, varname_len) == OV_STRCMP_EQUAL){
			if(ov_string_getlength(urlQuery->value[i]) == varname_len){
				//direct variablename in urlQuery
				Ov_SetDynamicVectorLength(re,re->veclen+1,STRING);
				ov_string_setvalue(&(re->value[re->veclen-1]), urlQuery->value[i+1]);
			}else{
				//match for varname and an opening bracket
				//if varname = var => search for var[ to allocate a list of variables
				if((urlQuery->value[i])[varname_len] == '['){
					Ov_SetDynamicVectorLength(re,re->veclen+1,STRING);
					ov_string_setvalue(&(re->value[re->veclen-1]), urlQuery->value[i+1]);
				}
			}
		}
	}
	return OV_ERR_OK;
}

/*
 * returns the format of the output
 */
static OV_RESULT extract_response_format(const OV_STRING_VEC* urlQuery, HTTP_RESPONSEFORMAT *response_format){
	OV_STRING_VEC match = {0,NULL};
	//output format
	kshttp_find_arguments(urlQuery, "format", &match);
	if(match.veclen>=1){
		if(ov_string_compare(match.value[0], "ksx") == OV_STRCMP_EQUAL){
			*response_format = KSX;
		}else if(ov_string_compare(match.value[0], "json") == OV_STRCMP_EQUAL){
			*response_format = KSJSON;
		}else if(ov_string_compare(match.value[0], "tcl") == OV_STRCMP_EQUAL){
			*response_format = KSTCL;
		}else if(ov_string_compare(match.value[0], "plain") == OV_STRCMP_EQUAL){
			*response_format = PLAIN;
		}
	}
	Ov_SetDynamicVectorLength(&match,0,STRING);
	return OV_ERR_OK;
}

#define PARSE_HTTP_HEADER_RETURN if(clientRequest->response_format == FORMATUNDEFINED){\
			clientRequest->response_format = FORMATDEFAULT;\
		}\
		ov_string_setvalue(&RequestLine, NULL);\
		ov_string_freelist(plist);\
		ov_string_freelist(pKeyValuepair);\
		return
/**
 * Parse raw http HTTP request into a get command and a list of arguments
 *
 * @param clientRequest
 * @param serverResponse
 * @return
 */
OV_RESULT kshttp_parse_http_header_from_client(HTTP_REQUEST *clientRequest, HTTP_RESPONSE *serverResponse)
{
	OV_STRING* pallheaderslist=NULL;
	OV_UINT allheaderscount = 0;
	OV_STRING* plist=NULL;
	OV_STRING* pKeyValuepair=NULL;
	OV_STRING RequestLine=NULL;
	OV_STRING temp=NULL;
	OV_BOOL httphostsend = FALSE;

	OV_UINT i = 0, len = 0, len1 = 0;
	char *endPtr = NULL;
	unsigned long int tempulong = 0;

	//initialize return vector properly
	Ov_SetDynamicVectorLength(&clientRequest->urlQuery, 0, STRING);

	pallheaderslist = ov_string_split(clientRequest->requestHeader, "\r\n", &allheaderscount);
	if(allheaderscount<=0){
		PARSE_HTTP_HEADER_RETURN OV_ERR_BADPARAM; //400
	}
	//split out the first line containing the Method
	ov_string_setvalue(&RequestLine, pallheaderslist[0]);
	//split out the actual method
	plist = ov_string_split(RequestLine, " ", &len);
	if(len!=3){
		PARSE_HTTP_HEADER_RETURN OV_ERR_BADPARAM; //400
	}
	ov_string_setvalue(&clientRequest->requestMethod, plist[0]);
	ov_string_setvalue(&RequestLine, plist[1]);
	//does the client use HTTP 1.0?
	if(ov_string_compare(plist[2], "HTTP/1.0") == OV_STRCMP_EQUAL){
		//if the server uses 1.0, use 1.0 as the response
		ov_string_setvalue(&clientRequest->httpVersion, "1.0");
		ov_string_setvalue(&serverResponse->httpVersion, "1.0");
		serverResponse->keepAlive = FALSE;
	}else if(ov_string_compare(plist[2], "HTTP/1.1") == OV_STRCMP_EQUAL){
		//if the server uses 1.1, use 1.1 as the response
		ov_string_setvalue(&clientRequest->httpVersion, "1.1");
		ov_string_setvalue(&serverResponse->httpVersion, "1.1");
		serverResponse->keepAlive = TRUE;
	}else{
		//with an unknown request version, answer with default HTTP version
		ov_string_setvalue(&clientRequest->httpVersion, NULL);
		ov_string_setvalue(&serverResponse->httpVersion, "1.1");
		serverResponse->keepAlive = TRUE;
	}

	//rawrequest contains the vars and urlQuery (raw)
	ov_string_freelist(plist);
	plist = ov_string_split(RequestLine, "?", &len);
	//get the urlPath, it contains the /-prefixed call
	ov_string_setvalue(&clientRequest->urlPath, plist[0]);
	//check if the request contains an ?
	if(len > 1) {
		//at least one part after a "?" -> split up the command

		ov_string_setvalue(&RequestLine, plist[1]);

		ov_string_freelist(plist);
		plist = ov_string_split(RequestLine, "&", &len);
		if(len > 0){
			Ov_SetDynamicVectorLength(&clientRequest->urlQuery,2*len,STRING);
			for(i = 0;i < len;i++){
				ov_string_freelist(pKeyValuepair);
				pKeyValuepair = ov_string_split(plist[i], "=", &len1);
				//varname=value
				if(len1==2){
					if(pKeyValuepair[0][0] == '\0'){
						PARSE_HTTP_HEADER_RETURN OV_ERR_BADPARAM; //400;
					}
					//urldecode key and value
					ov_memstack_lock();
					temp = url_decode(pKeyValuepair[0]);
					ov_string_setvalue(&clientRequest->urlQuery.value[2 * i], temp);

					temp = url_decode(pKeyValuepair[1]);
					ov_string_setvalue(&clientRequest->urlQuery.value[2 * i +1], temp);
					ov_memstack_unlock();
					temp = NULL;
				}else{
					PARSE_HTTP_HEADER_RETURN OV_ERR_BADPARAM; //400;
				}
			}
		}
	}

	//check all other headers
	for (i=1; i<allheaderscount; i++){
		if(ov_string_match(pallheaderslist[i], "?ccept-?ncoding:*") == TRUE){
			if(ov_string_match(pallheaderslist[i], "*gzip*") == TRUE){
				serverResponse->compressionGzip = TRUE;
			}
		}else if(ov_string_compare(pallheaderslist[i], "?onnection: close") == OV_STRCMP_EQUAL){
			//scan header for Connection: close - the default behavior is keep-alive
			serverResponse->keepAlive = FALSE;
		}else if(ov_string_match(pallheaderslist[i], "?ccept:*") == TRUE){
			ov_string_setvalue(&clientRequest->Accept, pallheaderslist[i]);
			if(ov_string_comparei(pallheaderslist[i], "Accept: text/plain") == OV_STRCMP_EQUAL){
				clientRequest->response_format = PLAIN;
			}else if(ov_string_comparei(pallheaderslist[i], "Accept: text/tcl") == OV_STRCMP_EQUAL){
				clientRequest->response_format = KSTCL;
			}else if(ov_string_comparei(pallheaderslist[i], "Accept: text/xml") == OV_STRCMP_EQUAL ||	//RFC3023: preferd if "readable by casual users"
					ov_string_comparei(pallheaderslist[i], "Accept: application/xml") == OV_STRCMP_EQUAL ||	//RFC3023: preferd if it is "unreadable by casual users"
					ov_string_comparei(pallheaderslist[i], "Accept: text/ksx") == OV_STRCMP_EQUAL){
				clientRequest->response_format = KSX;
			}else if(ov_string_comparei(pallheaderslist[i], "Accept: application/json") == OV_STRCMP_EQUAL){
				clientRequest->response_format = KSJSON;
			}
		}else if(ov_string_match(pallheaderslist[i], "?ost:*") == TRUE){
			ov_string_freelist(plist);
			plist = ov_string_split(pallheaderslist[i], "Host: ", &len);
			if(len == 1){
				//we have no result? perhaps the header was lowercase... second try
				ov_string_freelist(plist);
				plist = ov_string_split(pallheaderslist[i], "host: ", &len);
			}
			if(len > 1){
				ov_string_setvalue(&clientRequest->host, plist[1]);
				httphostsend = TRUE;
			}else{
				//empty Host header
				httphostsend = TRUE;
			}
		}else if(ov_string_match(pallheaderslist[i], "?ontent-?ength:*") == TRUE){
			ov_string_freelist(plist);
			plist = ov_string_split(pallheaderslist[i], "Content-Length: ", &len);
			if(len == 1){
				//we have no result? perhaps the header was lowercase... second try
				ov_string_freelist(plist);
				plist = ov_string_split(pallheaderslist[i], "content-length: ", &len);
			}
			if(len > 1){
				tempulong = strtoul(plist[1], &endPtr, 10);
				if (ERANGE != errno &&
					tempulong < OV_VL_MAXUINT &&
					endPtr != plist[1])
				{
					clientRequest->contentLength = (OV_UINT)tempulong;
				}
			}
		}else if(FALSE && ov_string_match(pallheaderslist[i], "Upgrade:*h2c*") == TRUE){
			//perhaps this is a test if we support HTTP/2
			//as we do not support this binary protocol we do not really test for the header
		}
	}
	ov_string_freelist(pallheaderslist);

	if(ov_string_compare(clientRequest->httpVersion, "1.1") == OV_STRCMP_EQUAL && httphostsend == FALSE){
		/* RFC 7230 Section 5.4: "A server MUST respond with a 400 (Bad Request) status
		 * code to any HTTP/1.1 request message that lacks a Host header field..." */
		PARSE_HTTP_HEADER_RETURN OV_ERR_BADPARAM; //400
	}
	//try setting format via url parameter
	extract_response_format(&clientRequest->urlQuery, &clientRequest->response_format);
	PARSE_HTTP_HEADER_RETURN OV_ERR_OK;
}



/**
*	Converts percent characters in ascii characters, but skips /
*	Note: the memory for the returned string is allocated on the memory
*	stack, use ov_memstack_lock()/unlock() outside of this function
*/
OV_DLLFNCEXPORT OV_STRING kshttp_ov_path_topercent_noslash (
				OV_STRING org
) {
	OV_STRING newstring;
	unsigned char *s,*d;
	unsigned int upper, lower;

	newstring = (OV_STRING) ov_memstack_alloc(ov_path_percentsize(org)+1);
	d = (unsigned char *)org;
	s = (unsigned char *)newstring;
	while (*d) {
		if (*d == '/'){
			//preserve Slash
			*s = *d;
		}else
		if (!ov_path_isvalidchar(*d)) {
			upper = (*d) >> 4;
			lower = (*d) % 16;
			s[0] = '%';
			s[1] = (unsigned char)upper + (upper < 10 ? '0' : '7');
			s[2] = (unsigned char)lower + (lower < 10 ? '0' : '7');
			s = s + 2;
		}
		else *s = *d;
		s++;
		d++;
	}
	*s = 0;
	return newstring;
}
