#pragma once

#include <linux/types.h>
#include <linux/unistd.h>

#define GTP_UDP_PORT 2152u //!< TS 29 281
#define GTP_FLAGS 0x30     //!< Version: GTPv1, Protocol Type: GTP, Others: 0

// TS 29 281 - Section 6 GTP-U Message Formats
// Table 6.1-1: Messages in GTP-U
#define GTPU_ECHO_REQUEST (1)
#define GTPU_ECHO_RESPONSE (2)
#define GTPU_ERROR_INDICATION (26)
#define GTPU_SUPPORTED_EXTENSION_HEADERS_NOTIFICATION (31)
#define GTPU_END_MARKER (254)
#define GTPU_G_PDU (255)

struct gtpuhdr
{
#if __BYTE_ORDER == __LITTLE_ENDIAN
    unsigned int pn : 1;
    unsigned int s : 1;
    unsigned int e : 1;
    unsigned int spare : 1;
    unsigned int pt : 1;
    unsigned int version : 3;
#elif __BYTE_ORDER == __BIG_ENDIAN
    unsigned int version : 3;
    unsigned int pt : 1;
    unsigned int spare : 1;
    unsigned int e : 1;
    unsigned int s : 1;
    unsigned int pn : 1;
#else
#error "Please fix <bits/endian.h>"
#endif
    // Message Type: This field indicates the type of GTP-U message.
    __u8 message_type;
    // Length: This field indicates the length in octets of the payload, i.e. the rest of the packet following the mandatory
    // part of the GTP header (that is the first 8 octets). The Sequence Number, the N-PDU Number or any Extension
    // headers shall be considered to be part of the payload, i.e. included in the length count.
    __u16 message_length;
    // Tunnel Endpoint Identifier (TEID): This field unambiguously identifies a tunnel endpoint in the receiving
    // GTP-U protocol entity. The receiving end side of a GTP tunnel locally assigns the TEID value the transmitting
    // side has to use. The TEID value shall be assigned in a non-predictable manner for PGW S5/S8/S2a/S2b
    // interfaces (see 3GPP TS 33.250 [32]). The TEID shall be used by the receiving entity to find the PDP context,
    // except for the following cases:
    // -) The Echo Request/Response and Supported Extension Headers notification messages, where the Tunnel
    //    Endpoint Identifier shall be set to all zeroes
    // -) The Error Indication message where the Tunnel Endpoint Identifier shall be set to all zeros.
    __u32 teid;

    /*The options start here. */
} __attribute__((packed));


 /* Optional word of GTP header, present if any of E, S, PN is set. */
 struct gtp_hdr_ext {
     __u16 sqn;       
     __u8 npdu;         
     __u8 next_ext;     
 } __attribute__((packed));