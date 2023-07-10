*** Settings ***
Library           KubeLibrary

*** Variables ***
${NAMESPACE}  open5gs
${POD_NAME}  .*gnb-ues*
@{COMMAND_REGISTER_UE}  nr-ue  -v  -c  ue.yaml  -n  1  -i  999700000000002
@{COMMAND_SEND_PING}  ping  -c1  -W1  -I  uesimtun1  172.17.0.1

*** Test Cases ***
Register UE
    Find Namespaced Pods by Pattern and Execute Command  ${POD_NAME}  ${NAMESPACE}  ${COMMAND_REGISTER_UE}
Test traffic
    Find Namespaced Pods by Pattern and Execute Command  ${POD_NAME}  ${NAMESPACE}  ${COMMAND_SEND_PING}

*** Keywords ***
Kubernetes API responds
    [Documentation]  Check if API response code is 200
    @{ping}=    k8s_api_ping
    Should Be Equal As integers    ${ping}[1]    200

Find Namespaced Pods by Pattern and Execute Command
    [Arguments]  ${POD_NAME}  ${NAMESPACE}  ${COMMAND}
    Kubernetes API responds
    ${pods}=     List Namespaced Pod By Pattern   ${POD_NAME}    ${NAMESPACE}

    FOR    ${pod}    IN    @{pods}
        ${pod_name} =    Set Variable    ${pod.metadata.name}
        Log     Running command ${COMMAND} in pod ${pod_name}
        ${result}=    Get Namespaced Pod Exec    ${pod_name}  ${NAMESPACE}  ${COMMAND}
        Log    ${result}
    END

