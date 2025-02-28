
/******************************************************************************
*
*   FILE
*   ----
*   DSUnregistrationServiceType1.c
*
*   History
*   -------
*   2018-05-14   File created
*
*******************************************************************************
*
*   This file is generated by the 'acplt_builder' command
*
******************************************************************************/


#ifndef OV_COMPILE_LIBRARY_DSServices
#define OV_COMPILE_LIBRARY_DSServices
#endif


#include "DSServices.h"
#include "libov/ov_macros.h"
#include "service_helper.h"

OV_DLLFNCEXPORT OV_RESULT DSServices_DSUnregistrationServiceType1_executeService(OV_INSTPTR_openAASDiscoveryServer_DSService pinst, const json_data JsonInput, OV_STRING *JsonOutput, OV_STRING *errorMessage) {
    /*    
    *   local variables
    */
	// Parsing Body
	OV_STRING_VEC tags;
	tags.value = NULL;
	tags.veclen = 0;
	Ov_SetDynamicVectorLength(&tags, 2, STRING);
	ov_string_setvalue(&tags.value[0], "componentID");
	ov_string_setvalue(&tags.value[1], "securityKey");
	OV_UINT_VEC tokenIndex;
	tokenIndex.value = NULL;
	tokenIndex.veclen = 0;
	Ov_SetDynamicVectorLength(&tokenIndex, 2, UINT);

	jsonGetTokenIndexByTags(tags, JsonInput, 1, &tokenIndex);

	OV_STRING componentID = NULL;
	jsonGetValueByToken(JsonInput.js, &JsonInput.token[tokenIndex.value[0]+1], &componentID);
	OV_STRING securityKey = NULL;
	jsonGetValueByToken(JsonInput.js, &JsonInput.token[tokenIndex.value[1]+1], &securityKey);

	if (pinst->v_DBWrapperUsed.veclen == 0){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not find DBWrapper Object");
		goto FINALIZE;
	}
	// check securityKey
	OV_RESULT resultOV = checkSecurityKey(pinst->v_DBWrapperUsed, componentID, securityKey);
	if (resultOV){
		ov_string_setvalue(errorMessage, "SecurityKey is not correct");
		goto FINALIZE;
	}
	// delete all data to componentID in Database
	// TODO: Extend to MultiDBWrapper
	OV_INSTPTR_openAASDiscoveryServer_DBWrapper pDBWrapper = NULL;
	OV_VTBLPTR_openAASDiscoveryServer_DBWrapper pDBWrapperVTable = NULL;
	pDBWrapper = Ov_DynamicPtrCast(openAASDiscoveryServer_DBWrapper, ov_path_getobjectpointer(pinst->v_DBWrapperUsed.value[0], 2));
	if (!pDBWrapper){
		ov_string_setvalue(errorMessage, "Internel Error");
		ov_logfile_error("Could not find DBWrapper Object");
		goto FINALIZE;
	}
	Ov_GetVTablePtr(openAASDiscoveryServer_DBWrapper,pDBWrapperVTable, pDBWrapper);
	OV_STRING tmpFields = "ComponentID";
	OV_STRING tmpValues = NULL;
	ov_string_print(&tmpValues, "'%s'", componentID);
	OV_STRING table  = "SecurityData";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in SecurityData in database");
		goto FINALIZE;
	}
	table  = "Endpoints";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Endpoints in database");
		goto FINALIZE;
	}
	table  = "statements_TextBoolean";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Statements in database");
		goto FINALIZE;
	}
	table  = "statements_Numeric";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Statements in database");
		goto FINALIZE;
	}
	table  = "CarrierID";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Statements in database");
		goto FINALIZE;
	}
	table  = "PropertyID";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Statements in database");
		goto FINALIZE;
	}
	table  = "ExpressionSemantic";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Statements in database");
		goto FINALIZE;
	}
	table  = "Relation";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Statements in database");
		goto FINALIZE;
	}
	table  = "SubModel";
	resultOV = pDBWrapperVTable->m_deleteData(pDBWrapper, table, &tmpFields, 1, &tmpValues, 1);
	if (resultOV){
		ov_string_setvalue(errorMessage, "Internal Error");
		ov_logfile_error("Could not delete data in Statements in database");
		goto FINALIZE;
	}

	FINALIZE:
	ov_string_setvalue(&tmpValues, NULL);
	ov_string_print(JsonOutput, "\"body\":{}");
	Ov_SetDynamicVectorLength(&tags, 0, STRING);
	Ov_SetDynamicVectorLength(&tokenIndex, 0, UINT);
	ov_string_setvalue(&componentID, NULL);
	ov_string_setvalue(&securityKey, NULL);
    return OV_ERR_OK;
}

OV_DLLFNCEXPORT OV_ACCESS DSServices_DSUnregistrationServiceType1_getaccess(
	OV_INSTPTR_ov_object	pobj,
	const OV_ELEMENT		*pelem,
	const OV_TICKET			*pticket
) {
    /*    
    *   local variables
    */

	switch(pelem->elemtype) {
		case OV_ET_VARIABLE:
			if(pelem->elemunion.pvar->v_offset >= offsetof(OV_INST_ov_object,__classinfo)) {
				if(pelem->elemunion.pvar->v_vartype == OV_VT_CTYPE)
					return OV_AC_NONE;
				else{
					if(pelem->elemunion.pvar->v_flags == 256) { // InputFlag is set
						return OV_AC_READWRITE;
					}
					/* Nicht FB? */
					if(pelem->elemunion.pvar->v_varprops & OV_VP_SETACCESSOR) {
						return OV_AC_READWRITE;
					}
					return OV_AC_READ;
				}
			}
		break;
		default:
		break;
	}

	return ov_object_getaccess(pobj, pelem, pticket);
}

