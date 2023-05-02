# Possible cause codes

| Cause Name                           | Value (uint8) | Possible causes                                         |
|--------------------------------------|---------------|---------------------------------------------------------|
| CauseRequestAccepted                 | 1             | Request successfully processed                          |
| CauseRequestRejected                 | 64            | Encountered unknown error                               |
| CauseSessionContextNotFound          | 65            | Trying to delete or modify session that does not exists |
| CauseMandatoryIEMissing              | 66            | Got message without NodeID or other mandatory IE        |
| CauseConditionalIEMissing            | 67            | Currently not in use                                    |
| CauseInvalidLength                   | 68            | Currently not in use                                    |
| CauseMandatoryIEIncorrect            | 69            | Currently not in use                                    |
| CauseInvalidForwardingPolicy         | 70            | Currently not in use                                    |
| CauseInvalidFTEIDAllocationOption    | 71            | Currently not in use                                    |
| CauseNoEstablishedPFCPAssociation    | 72            | Trying to create session, but no association is present |
| CauseRuleCreationModificationFailure | 73            | There was an error in applying session rules            |
| CausePFCPEntityInCongestion          | 74            | Currently not in use                                    |
| CauseNoResourcesAvailable            | 75            | Currently not in use                                    |
| CauseServiceNotSupported             | 76            | Currently not in use                                    |
| CauseSystemFailure                   | 77            | Currently not in use                                    |
| CauseRedirectionRequested            | 78            | Currently not in use                                    |
