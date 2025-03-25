*** Settings ***
Library    RequestsLibrary
Library    Collections
Library    Process
Library    OperatingSystem
Library    String

*** Variables ***
${EUPF_API_ADDRESS}    http://localhost:8080
${TCPREPLAY_LIMIT}    7000000
${TCPREPLAY_THREADS}  6
${INTERFACE}          lo
${PAYLOAD}            ${EMPTY}

*** Test Cases ***
Perform Load Test
    [Setup]    Check API Availability
    Generate Payload

    ${EUPF_PROCESSED_ON_START}    Get EUPF Packet Processed    ${EUPF_API_ADDRESS}

    Set eUPF Uplink PDR    teid=0    far_id=0    qer_id=0
    Set eUPF FAR    far_id=0
    Set eUPF QER    qer_id=0

    ${PACKET_PATH}  Create GTP-U Packet And Save To File
    ${TCPREPLAY_EXTRA_ARGS}   Create List    --limit=${TCPREPLAY_LIMIT}

    ${handles}    Create List
    FOR    ${i}    IN RANGE    ${TCPREPLAY_THREADS}
        ${args}    Create List    -i    ${INTERFACE}    -t    @{TCPREPLAY_EXTRA_ARGS}    ${PACKET_PATH}
        ${handle}=    Start Process    tcpreplay    @{args}
        Append To List    ${handles}    ${handle}
    END

    FOR    ${handle}    IN    @{handles}
        Wait For Process    ${handle}    timeout=900
    END

    ${RESULT}    Create Dictionary    pps=0    mbps=0    packets=0
    Log To Console    Resulting pps: ${RESULT}[pps]
    Log To Console    Resulting mbps: ${RESULT}[mbps]
    Log To Console    Resulting packets: ${RESULT}[packets]

    ${TCPREPLAY_SENT}    Set Variable    ${RESULT}[packets]
    ${EUPF_PROCESSED_ON_STOP}    Get EUPF Packet Processed    ${EUPF_API_ADDRESS}

    Should Be True    ${EUPF_PROCESSED_ON_STOP} == ${EUPF_PROCESSED_ON_START}
    ${EUPF_PROCESSED}    Evaluate    ${EUPF_PROCESSED_ON_STOP} - ${EUPF_PROCESSED_ON_START}

    Should Be Equal As Integers    ${TCPREPLAY_SENT}    ${EUPF_PROCESSED}    msg=Sent packets count doesn't equal processed packets

*** Keywords ***
Check API Availability
    ${response}    Run Keyword And Return Status    GET    ${EUPF_API_ADDRESS}/api/v1/xdp_stats
    Run Keyword If    not ${response}    Fail    API at ${EUPF_API_ADDRESS} is not available!

Generate Payload
    ${PAYLOAD}    Evaluate    'a' * 1024
    Set Global Variable    ${PAYLOAD}

Get EUPF Packet Processed
    [Arguments]    ${API_ADDRESS}
    ${response}    GET  ${API_ADDRESS}/api/v1/xdp_stats  expected_status=200
    ${tx}    Set Variable    ${response.json()}[tx]
    ${redirect}    Set Variable    ${response.json()}[redirect]
    ${result}    Evaluate    ${tx} + ${redirect}
    RETURN    ${result}

Create GTP-U Packet And Save To File
    ${PYTHON_SCRIPT}    Catenate
    ...    from scapy.all import Ether, IP, UDP, Raw, wrpcap;
    ...    packet = Ether()/IP(src='127.0.0.1', dst='127.0.0.1')/UDP(sport=2152, dport=2152)/Raw(load='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa');
    ...    wrpcap('gtp_packet.pcap', [packet])

    Run Process    python3 -c "${PYTHON_SCRIPT}"    shell=True

    RETURN    gtp_packet.pcap

Set eUPF Uplink PDR
    [Arguments]   ${teid}=0    ${far_id}=0    ${qer_id}=0

    # forced casting of data to integer type
    ${teid}    Convert To Integer    ${teid}
    ${far_id}    Convert To Integer    ${far_id}
    ${qer_id}                    Convert To Integer    ${qer_id}
    ${outer_header_removal}    Convert To Integer    ${0}

    ${body}    Create Dictionary    
    ...    id=${teid}    
    ...    outer_header_removal=${outer_header_removal}    
    ...    far_id=${far_id}    
    ...    qer_id=${qer_id}

    ${response}    
    ...    PUT    
    ...    ${EUPF_API_ADDRESS}/api/v1/uplink_pdr_map/${teid}    
    ...    json=${body}

    Log    ${response.json()}

Set eUPF FAR
    [Arguments]   ${far_id}=0

    ${far_id}    Convert To Integer    ${far_id}
    ${teid}    Convert To Integer    ${0}
    ${remote_ip}    Convert To Integer    ${0}
    ${local_ip}    Convert To Integer    ${0}
    ${transport_level_marking}    Convert To Integer    ${0}
    ${action}    Convert To Integer    ${2}
    ${outer_header_creation}    Convert To Integer    ${0}

    ${body}    Create Dictionary    
    ...    action=${action}     
    ...    outer_header_creation=${outer_header_creation}     
    ...    teid=${teid}    
    ...    remote_ip=${remote_ip}    
    ...    local_ip=${local_ip}    
    ...    transport_level_marking=${transport_level_marking}

    ${response}    
    ...    PUT    
    ...    ${EUPF_API_ADDRESS}/api/v1/far_map/${far_id}    
    ...    json=${body}

    Log    ${response.json()}

Set eUPF QER
    [Arguments]   ${qer_id}=0

    ${qer_id}    Convert To Integer    ${qer_id}
    ${gate_status_ul}    Convert To Integer    ${0}
    ${gate_status_dl}    Convert To Integer    ${0}
    ${qfi}    Convert To Integer    ${0}
    ${max_bitrate_ul}    Convert To Integer    ${0}
    ${max_bitrate_dl}    Convert To Integer    ${0}

    ${body}    Create Dictionary
    ...    gate_status_ul=${gate_status_ul}    
    ...    gate_status_dl=${gate_status_dl}    
    ...    qfi=${qfi}    
    ...    max_bitrate_ul=${max_bitrate_ul}    
    ...    max_bitrate_dl=${max_bitrate_dl}

    ${response}    
    ...    PUT    
    ...    ${EUPF_API_ADDRESS}/api/v1/qer_map/${qer_id}
    ...    json=${body}

    Log    ${response.json()}
