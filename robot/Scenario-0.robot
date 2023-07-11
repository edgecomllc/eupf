*** Settings ***
Library           KubeLibrary

*** Variables ***
${NAMESPACE}  open5gs
${POD_NAME}  .*gnb-ues*
@{COMMAND_CHECK_PDU}  nr-cli  imsi-999700000000001  -e  status
@{COMMAND_SEND_PING}  ping  -c1  -W1  -I  uesimtun0  172.17.0.1

*** Test Cases ***
PDU Session is up and running
    Check UE status
    Test traffic

*** Keywords ***
Find Namespaced Pods by Pattern and Execute Command
    [Documentation]  This keyword expects pattern to match only one pod
    [Arguments]  ${POD_NAME}  ${NAMESPACE}  ${COMMAND}
    ${pods}=     List Namespaced Pod By Pattern   ${POD_NAME}    ${NAMESPACE}
    FOR    ${pod}    IN    @{pods}
        ${pod_name} =    Set Variable    ${pod.metadata.name}
        Log     Running command ${COMMAND} in pod ${pod_name}
        ${result}=    Get Namespaced Pod Exec    ${pod_name}  ${NAMESPACE}  ${COMMAND}
    END
    [Return]  ${result}

Check UE status
    ${out}=    Find Namespaced Pods by Pattern and Execute Command  ${POD_NAME}  ${NAMESPACE}  ${COMMAND_CHECK_PDU}
    Should Be True      "cm-state: CM-CONNECTED" in """${out}"""
    Should Be True      "rm-state: RM-REGISTERED" in """${out}"""
    Should Be True      "mm-state: MM-REGISTERED" in """${out}"""

Test traffic
    ${out}=    Find Namespaced Pods by Pattern and Execute Command  ${POD_NAME}  ${NAMESPACE}  ${COMMAND_SEND_PING}
    Should Be True      "0% packet loss" in """${out}"""