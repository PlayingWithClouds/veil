export default {
    "scalars": [
        1,
        2,
        3,
        4,
        5,
        14,
        16,
        17,
        18,
        22,
        30,
        41,
        44,
        61,
        63,
        74,
        75,
        87,
        88,
        95,
        104,
        109,
        112,
        121,
        122,
        124,
        126,
        128,
        129,
        130,
        131,
        137,
        141
    ],
    "types": {
        "Query": {
            "play": [
                7,
                {
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "follows": [
                10
            ],
            "follow": [
                10,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "followOf": [
                10,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "newSince": [
                134,
                {
                    "followId": [
                        2,
                        "ID!"
                    ],
                    "since": [
                        1
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "notifications": [
                13,
                {
                    "unreadOnly": [
                        4
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "unreadNotificationCount": [
                3
            ],
            "notificationChannels": [
                15
            ],
            "limits": [
                19
            ],
            "limitStatus": [
                20
            ],
            "usage": [
                21,
                {
                    "window": [
                        17,
                        "LimitWindow!"
                    ]
                }
            ],
            "automations": [
                27
            ],
            "automation": [
                27,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "automationRuns": [
                31,
                {
                    "automationId": [
                        2
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "automationRun": [
                31,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "automationActions": [
                1
            ],
            "camRooms": [
                34,
                {
                    "query": [
                        33
                    ]
                }
            ],
            "camModels": [
                35,
                {
                    "live": [
                        4
                    ],
                    "autoRecord": [
                        4
                    ],
                    "performerId": [
                        2
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "camModel": [
                35,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "activeRecordings": [
                37
            ],
            "categories": [
                39
            ],
            "category": [
                39,
                {
                    "id": [
                        2
                    ],
                    "slug": [
                        1
                    ]
                }
            ],
            "collections": [
                42
            ],
            "collection": [
                42,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "builtInCollection": [
                42,
                {
                    "kind": [
                        41,
                        "CollectionKind!"
                    ]
                }
            ],
            "collectionsWith": [
                42,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "fieldDefs": [
                45,
                {
                    "kind": [
                        129
                    ]
                }
            ],
            "recommendations": [
                48,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "related": [
                48,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "shorts": [
                50,
                {
                    "filter": [
                        147
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "taste": [
                51
            ],
            "bestMoment": [
                5,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "alikes": [
                55,
                {
                    "kind": [
                        129,
                        "RecordKind!"
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "libraryRoots": [
                56
            ],
            "libraryRoot": [
                56,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "planOrganize": [
                57,
                {
                    "rootId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "downloads": [
                60
            ],
            "sameMedia": [
                134,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "nearest": [
                48,
                {
                    "kind": [
                        63,
                        "EmbeddingKind!"
                    ],
                    "model": [
                        1,
                        "String!"
                    ],
                    "vector": [
                        5,
                        "[Float!]!"
                    ],
                    "kinds": [
                        129,
                        "[RecordKind!]"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "similar": [
                48,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "kind": [
                        63,
                        "EmbeddingKind!"
                    ],
                    "model": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "videos": [
                38,
                {
                    "filter": [
                        147
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "videoCount": [
                3,
                {
                    "filter": [
                        147
                    ]
                }
            ],
            "video": [
                38,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "galleries": [
                47,
                {
                    "filter": [
                        147
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "galleryCount": [
                3,
                {
                    "filter": [
                        147
                    ]
                }
            ],
            "gallery": [
                47,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "performers": [
                81,
                {
                    "text": [
                        1
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "performer": [
                81,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "studios": [
                83,
                {
                    "text": [
                        1
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "studio": [
                83,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "tags": [
                84,
                {
                    "text": [
                        1
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "tag": [
                84,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "media": [
                134,
                {
                    "filter": [
                        147
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "mediaCount": [
                3,
                {
                    "filter": [
                        147
                    ]
                }
            ],
            "audios": [
                66,
                {
                    "filter": [
                        147
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "audioCount": [
                3,
                {
                    "filter": [
                        147
                    ]
                }
            ],
            "audio": [
                66,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "stories": [
                89,
                {
                    "filter": [
                        147
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "storyCount": [
                3,
                {
                    "filter": [
                        147
                    ]
                }
            ],
            "story": [
                89,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "posts": [
                91,
                {
                    "filter": [
                        147
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "postCount": [
                3,
                {
                    "filter": [
                        147
                    ]
                }
            ],
            "post": [
                91,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "groups": [
                92,
                {
                    "text": [
                        1
                    ],
                    "videoId": [
                        2
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "group": [
                92,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "settings": [
                128
            ],
            "apiTokens": [
                93
            ],
            "accounts": [
                96
            ],
            "contributions": [
                97,
                {
                    "kind": [
                        1
                    ]
                }
            ],
            "pluginView": [
                98,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "view": [
                        1,
                        "String!"
                    ],
                    "input": [
                        128
                    ]
                }
            ],
            "clips": [
                107,
                {
                    "videoId": [
                        2
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "clip": [
                107,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                103
            ],
            "plugin": [
                103,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "pluginPackages": [
                116,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "trash": [
                120,
                {
                    "kind": [
                        129,
                        "RecordKind!"
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "jobs": [
                123,
                {
                    "statuses": [
                        122,
                        "[JobStatus!]"
                    ],
                    "kinds": [
                        121,
                        "[JobKind!]"
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "job": [
                123,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "devices": [
                125
            ],
            "currentDevice": [
                125
            ],
            "storageLocations": [
                127,
                {
                    "deviceId": [
                        2
                    ]
                }
            ],
            "userStates": [
                135,
                {
                    "filter": [
                        136
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "history": [
                135,
                {
                    "kind": [
                        129
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "events": [
                138,
                {
                    "filter": [
                        140
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "eventCount": [
                3,
                {
                    "filter": [
                        140
                    ]
                }
            ],
            "recentSearches": [
                1,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "savedFilters": [
                148
            ],
            "savedFilter": [
                148,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "String": {},
        "ID": {},
        "Int": {},
        "Boolean": {},
        "Float": {},
        "Mutation": {
            "scrape": [
                8,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "followRecord": [
                10,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "intervalHours": [
                        3
                    ],
                    "notify": [
                        4
                    ]
                }
            ],
            "followSearch": [
                10,
                {
                    "text": [
                        1,
                        "String!"
                    ],
                    "intervalHours": [
                        3
                    ],
                    "notify": [
                        4
                    ]
                }
            ],
            "createFollow": [
                10,
                {
                    "input": [
                        11,
                        "FollowInput!"
                    ]
                }
            ],
            "updateFollow": [
                10,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        12,
                        "FollowEditInput!"
                    ]
                }
            ],
            "unfollow": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "runFollow": [
                123,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "markNotificationsRead": [
                4,
                {
                    "ids": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "markAllNotificationsRead": [
                4
            ],
            "createNotificationChannel": [
                15,
                {
                    "kind": [
                        14,
                        "ChannelKind!"
                    ],
                    "name": [
                        1,
                        "String!"
                    ],
                    "config": [
                        128
                    ],
                    "enabled": [
                        4
                    ]
                }
            ],
            "updateNotificationChannel": [
                15,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "name": [
                        1
                    ],
                    "config": [
                        128
                    ],
                    "enabled": [
                        4
                    ]
                }
            ],
            "deleteNotificationChannel": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "createLimit": [
                19,
                {
                    "kind": [
                        16,
                        "LimitKind!"
                    ],
                    "window": [
                        17,
                        "LimitWindow!"
                    ],
                    "amount": [
                        3,
                        "Int!"
                    ],
                    "action": [
                        18,
                        "LimitAction!"
                    ],
                    "enabled": [
                        4
                    ]
                }
            ],
            "updateLimit": [
                19,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "amount": [
                        3
                    ],
                    "action": [
                        18
                    ],
                    "enabled": [
                        4
                    ]
                }
            ],
            "deleteLimit": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "createAutomation": [
                27,
                {
                    "input": [
                        28,
                        "AutomationInput!"
                    ]
                }
            ],
            "updateAutomation": [
                27,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        29,
                        "AutomationEditInput!"
                    ]
                }
            ],
            "deleteAutomation": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "runAutomation": [
                31,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "subjects": [
                        132,
                        "[RecordRefInput!]"
                    ],
                    "dryRun": [
                        4
                    ]
                }
            ],
            "undoAutomationRun": [
                31,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "watchCam": [
                7,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "setCamAutoRecord": [
                35,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "on": [
                        4,
                        "Boolean!"
                    ],
                    "quality": [
                        1
                    ]
                }
            ],
            "linkCamPerformer": [
                35,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "performerId": [
                        2
                    ]
                }
            ],
            "startRecording": [
                123,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "quality": [
                        1
                    ],
                    "fromSecondsAgo": [
                        3
                    ]
                }
            ],
            "stopRecording": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "createCategory": [
                39,
                {
                    "input": [
                        40,
                        "CategoryInput!"
                    ]
                }
            ],
            "editCategory": [
                39,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        40,
                        "CategoryInput!"
                    ]
                }
            ],
            "reorderCategories": [
                39,
                {
                    "ids": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "createCollection": [
                42,
                {
                    "name": [
                        1,
                        "String!"
                    ],
                    "details": [
                        1
                    ]
                }
            ],
            "editCollection": [
                42,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "name": [
                        1
                    ],
                    "details": [
                        1
                    ],
                    "coverUri": [
                        1
                    ]
                }
            ],
            "addToCollection": [
                42,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subjects": [
                        132,
                        "[RecordRefInput!]!"
                    ],
                    "position": [
                        3
                    ]
                }
            ],
            "removeFromCollection": [
                42,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subjects": [
                        132,
                        "[RecordRefInput!]!"
                    ]
                }
            ],
            "moveInCollection": [
                42,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "position": [
                        3,
                        "Int!"
                    ]
                }
            ],
            "clearCollection": [
                42,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "saveFieldDef": [
                45,
                {
                    "id": [
                        2
                    ],
                    "input": [
                        46,
                        "FieldDefInput!"
                    ]
                }
            ],
            "deleteFieldDef": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "setFieldValue": [
                135,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "fieldDefId": [
                        2,
                        "ID!"
                    ],
                    "value": [
                        128
                    ]
                }
            ],
            "lookup": [
                54,
                {
                    "input": [
                        53,
                        "LookupInput!"
                    ]
                }
            ],
            "merge": [
                134,
                {
                    "kind": [
                        129,
                        "RecordKind!"
                    ],
                    "keepId": [
                        2,
                        "ID!"
                    ],
                    "mergeId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "mergeAlike": [
                134,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "dismissAlike": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addLibraryRoot": [
                56,
                {
                    "locationId": [
                        2,
                        "ID!"
                    ],
                    "path": [
                        1,
                        "String!"
                    ],
                    "scanIntervalMinutes": [
                        3
                    ],
                    "organizeTemplate": [
                        1
                    ]
                }
            ],
            "updateLibraryRoot": [
                56,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "scanIntervalMinutes": [
                        3
                    ],
                    "organizeTemplate": [
                        1
                    ]
                }
            ],
            "removeLibraryRoot": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "scanLibraryRoot": [
                123,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "applyOrganize": [
                3,
                {
                    "rootId": [
                        2,
                        "ID!"
                    ],
                    "moves": [
                        58,
                        "[FileMoveInput!]!"
                    ]
                }
            ],
            "trashFile": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "download": [
                123,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "streamId": [
                        2
                    ]
                }
            ],
            "deleteDownload": [
                4,
                {
                    "jobId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addFingerprint": [
                62,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "fileId": [
                        2
                    ],
                    "algorithm": [
                        61,
                        "FingerprintAlgorithm!"
                    ],
                    "value": [
                        1,
                        "String!"
                    ],
                    "durationSeconds": [
                        5
                    ]
                }
            ],
            "fingerprintFile": [
                123,
                {
                    "fileId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "snapshotSource": [
                123,
                {
                    "sourceId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addEmbedding": [
                64,
                {
                    "input": [
                        65,
                        "EmbeddingInput!"
                    ]
                }
            ],
            "removeEmbedding": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "fetchVideo": [
                38,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "fetchGallery": [
                47,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "browse": [
                76,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "kind": [
                        75,
                        "ListKind!"
                    ],
                    "query": [
                        1
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "search": [
                76,
                {
                    "kind": [
                        75,
                        "ListKind!"
                    ],
                    "query": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "editVideo": [
                38,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        68,
                        "MediaEdit!"
                    ]
                }
            ],
            "editGallery": [
                47,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        68,
                        "MediaEdit!"
                    ]
                }
            ],
            "editPerformer": [
                81,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        69,
                        "PerformerEdit!"
                    ]
                }
            ],
            "editStudio": [
                83,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        70,
                        "StudioEdit!"
                    ]
                }
            ],
            "editTag": [
                84,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        71,
                        "TagEdit!"
                    ]
                }
            ],
            "lockField": [
                72,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "field": [
                        1,
                        "String!"
                    ],
                    "locked": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "revertRevision": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "editMedia": [
                134,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "edit": [
                        68,
                        "MediaEdit!"
                    ]
                }
            ],
            "fetchMedia": [
                134,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "readChapter": [
                90,
                {
                    "storyId": [
                        2,
                        "ID!"
                    ],
                    "position": [
                        3,
                        "Int!"
                    ]
                }
            ],
            "updateSettings": [
                128,
                {
                    "values": [
                        128,
                        "JSON!"
                    ]
                }
            ],
            "createApiToken": [
                94,
                {
                    "name": [
                        1,
                        "String!"
                    ],
                    "scopes": [
                        1,
                        "[String!]!"
                    ],
                    "expiresAt": [
                        1
                    ]
                }
            ],
            "revokeApiToken": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "loginAccount": [
                96,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "values": [
                        128,
                        "JSON!"
                    ]
                }
            ],
            "checkAccount": [
                96,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "logoutAccount": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "submitPluginView": [
                102,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "view": [
                        1,
                        "String!"
                    ],
                    "action": [
                        1,
                        "String!"
                    ],
                    "values": [
                        128
                    ],
                    "input": [
                        128
                    ]
                }
            ],
            "addSegment": [
                105,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "input": [
                        106,
                        "SegmentInput!"
                    ]
                }
            ],
            "editSegment": [
                105,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        106,
                        "SegmentInput!"
                    ]
                }
            ],
            "addClip": [
                107,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        108,
                        "ClipInput!"
                    ]
                }
            ],
            "editClip": [
                107,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        108,
                        "ClipInput!"
                    ]
                }
            ],
            "attachClipFile": [
                107,
                {
                    "clipId": [
                        2,
                        "ID!"
                    ],
                    "fileId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addSubtitle": [
                110,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "input": [
                        111,
                        "SubtitleInput!"
                    ]
                }
            ],
            "removeSubtitle": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addScript": [
                113,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "uri": [
                        1,
                        "String!"
                    ],
                    "axis": [
                        1
                    ]
                }
            ],
            "removeScript": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "setTrickplay": [
                38,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "trickplay": [
                        114
                    ]
                }
            ],
            "reportStreamSpeed": [
                85,
                {
                    "streamId": [
                        2,
                        "ID!"
                    ],
                    "kilobitsPerSecond": [
                        3,
                        "Int!"
                    ]
                }
            ],
            "refreshSources": [
                134,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "setPluginEnabled": [
                4,
                {
                    "name": [
                        1,
                        "String!"
                    ],
                    "enabled": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "updatePluginSettings": [
                4,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "values": [
                        119,
                        "[PluginSettingValueInput!]!"
                    ]
                }
            ],
            "installPlugin": [
                4,
                {
                    "packageName": [
                        1,
                        "String!"
                    ]
                }
            ],
            "uninstallPlugin": [
                4,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "updatePlugins": [
                1
            ],
            "trashRecord": [
                4,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "restoreRecord": [
                4,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "purgeRecord": [
                4,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "emptyTrash": [
                3
            ],
            "queueFetch": [
                123,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ]
                }
            ],
            "cancelJob": [
                123,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "retryJob": [
                123,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "clearFinishedJobs": [
                3
            ],
            "renameDevice": [
                125,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "addStorageLocation": [
                127,
                {
                    "kind": [
                        126,
                        "StorageKind!"
                    ],
                    "uri": [
                        1,
                        "String!"
                    ],
                    "quotaBytes": [
                        3
                    ]
                }
            ],
            "setStorageQuota": [
                127,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "quotaBytes": [
                        3
                    ]
                }
            ],
            "removeStorageLocation": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "rate": [
                135,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "rating": [
                        5
                    ]
                }
            ],
            "setFavorite": [
                135,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "favorite": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setBlocked": [
                135,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "blocked": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setNotes": [
                135,
                {
                    "subject": [
                        132,
                        "RecordRefInput!"
                    ],
                    "notes": [
                        1,
                        "String!"
                    ]
                }
            ],
            "recordEvent": [
                138,
                {
                    "event": [
                        139,
                        "EventInput!"
                    ]
                }
            ],
            "clearEvents": [
                3,
                {
                    "kinds": [
                        137,
                        "[EventKind!]"
                    ]
                }
            ],
            "saveFilter": [
                148,
                {
                    "id": [
                        2
                    ],
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        147,
                        "FilterInput!"
                    ],
                    "pinned": [
                        4
                    ]
                }
            ],
            "deleteSavedFilter": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "Playable": {
            "url": [
                1
            ],
            "mimeType": [
                1
            ],
            "quality": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Scraped": {
            "kind": [
                1
            ],
            "id": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Subscription": {
            "notificationAdded": [
                13
            ],
            "__typename": [
                1
            ]
        },
        "Follow": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "filter": [
                146
            ],
            "targetKind": [
                129
            ],
            "targetId": [
                2
            ],
            "target": [
                134
            ],
            "intervalHours": [
                3
            ],
            "enabled": [
                4
            ],
            "notify": [
                4
            ],
            "lastRunAt": [
                1
            ],
            "nextRunAt": [
                1
            ],
            "lastError": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FollowInput": {
            "name": [
                1
            ],
            "filter": [
                147
            ],
            "intervalHours": [
                3
            ],
            "notify": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "FollowEditInput": {
            "name": [
                1
            ],
            "filter": [
                147
            ],
            "intervalHours": [
                3
            ],
            "enabled": [
                4
            ],
            "notify": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "Notification": {
            "id": [
                2
            ],
            "at": [
                1
            ],
            "kind": [
                1
            ],
            "title": [
                1
            ],
            "body": [
                1
            ],
            "subject": [
                133
            ],
            "record": [
                134
            ],
            "read": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "ChannelKind": {},
        "NotificationChannel": {
            "id": [
                2
            ],
            "kind": [
                14
            ],
            "name": [
                1
            ],
            "config": [
                128
            ],
            "enabled": [
                4
            ],
            "delivers": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "LimitKind": {},
        "LimitWindow": {},
        "LimitAction": {},
        "Limit": {
            "id": [
                2
            ],
            "kind": [
                16
            ],
            "window": [
                17
            ],
            "amount": [
                3
            ],
            "action": [
                18
            ],
            "enabled": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "LimitStatus": {
            "limit": [
                19
            ],
            "used": [
                5
            ],
            "reached": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "Usage": {
            "since": [
                1
            ],
            "until": [
                1
            ],
            "watchMinutes": [
                5
            ],
            "sessions": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "TriggerKind": {},
        "Trigger": {
            "kind": [
                22
            ],
            "cron": [
                1
            ],
            "eventKinds": [
                1
            ],
            "filter": [
                146
            ],
            "pluginTrigger": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "TriggerInput": {
            "kind": [
                22
            ],
            "cron": [
                1
            ],
            "eventKinds": [
                1
            ],
            "filter": [
                147
            ],
            "pluginTrigger": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Step": {
            "action": [
                1
            ],
            "args": [
                128
            ],
            "condition": [
                146
            ],
            "continueOnError": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "StepInput": {
            "action": [
                1
            ],
            "args": [
                128
            ],
            "condition": [
                147
            ],
            "continueOnError": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "Automation": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "enabled": [
                4
            ],
            "trigger": [
                23
            ],
            "steps": [
                25
            ],
            "builtIn": [
                4
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "AutomationInput": {
            "name": [
                1
            ],
            "enabled": [
                4
            ],
            "trigger": [
                24
            ],
            "steps": [
                26
            ],
            "__typename": [
                1
            ]
        },
        "AutomationEditInput": {
            "name": [
                1
            ],
            "enabled": [
                4
            ],
            "trigger": [
                24
            ],
            "steps": [
                26
            ],
            "__typename": [
                1
            ]
        },
        "RunStatus": {},
        "AutomationRun": {
            "id": [
                2
            ],
            "automation": [
                27
            ],
            "startedAt": [
                1
            ],
            "finishedAt": [
                1
            ],
            "status": [
                30
            ],
            "dryRun": [
                4
            ],
            "triggerData": [
                128
            ],
            "steps": [
                32
            ],
            "__typename": [
                1
            ]
        },
        "StepResult": {
            "action": [
                1
            ],
            "ok": [
                4
            ],
            "output": [
                128
            ],
            "error": [
                1
            ],
            "touched": [
                133
            ],
            "__typename": [
                1
            ]
        },
        "CamQuery": {
            "plugins": [
                1
            ],
            "text": [
                1
            ],
            "tag": [
                1
            ],
            "gender": [
                1
            ],
            "limit": [
                3
            ],
            "offset": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "CamRoomPage": {
            "rooms": [
                35
            ],
            "total": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "CamModel": {
            "id": [
                2
            ],
            "plugin": [
                1
            ],
            "username": [
                1
            ],
            "displayName": [
                1
            ],
            "performer": [
                81
            ],
            "thumbnail": [
                77
            ],
            "tags": [
                1
            ],
            "gender": [
                1
            ],
            "live": [
                36
            ],
            "lastSeenOnlineAt": [
                1
            ],
            "autoRecord": [
                4
            ],
            "recordQuality": [
                1
            ],
            "recordings": [
                38
            ],
            "userState": [
                135
            ],
            "__typename": [
                1
            ]
        },
        "CamLive": {
            "viewers": [
                3
            ],
            "subject": [
                1
            ],
            "startedAt": [
                1
            ],
            "thumbnail": [
                77
            ],
            "__typename": [
                1
            ]
        },
        "ActiveRecording": {
            "job": [
                123
            ],
            "camModel": [
                35
            ],
            "startedAt": [
                1
            ],
            "durationSeconds": [
                5
            ],
            "sizeBytes": [
                3
            ],
            "quality": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Video": {
            "recordedFrom": [
                35
            ],
            "categories": [
                39
            ],
            "related": [
                48,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "files": [
                59
            ],
            "fingerprints": [
                62
            ],
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                77
            ],
            "status": [
                131
            ],
            "studio": [
                83
            ],
            "performers": [
                81
            ],
            "tags": [
                84
            ],
            "durationSeconds": [
                5
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "vr": [
                78
            ],
            "preview": [
                77
            ],
            "trickplay": [
                79
            ],
            "gallery": [
                47
            ],
            "sources": [
                67
            ],
            "streams": [
                85
            ],
            "userState": [
                135
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                72
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "groups": [
                92
            ],
            "segments": [
                105
            ],
            "subtitles": [
                110
            ],
            "scripts": [
                113
            ],
            "clips": [
                107
            ],
            "comments": [
                115
            ],
            "__typename": [
                1
            ]
        },
        "Category": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "slug": [
                1
            ],
            "parent": [
                39
            ],
            "children": [
                39
            ],
            "matchTags": [
                84
            ],
            "cover": [
                77
            ],
            "position": [
                3
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "CategoryInput": {
            "name": [
                1
            ],
            "slug": [
                1
            ],
            "parentId": [
                2
            ],
            "matchTagIds": [
                2
            ],
            "coverUri": [
                1
            ],
            "position": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "CollectionKind": {},
        "Collection": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "kind": [
                41
            ],
            "details": [
                1
            ],
            "cover": [
                77
            ],
            "itemCount": [
                3
            ],
            "items": [
                43,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "CollectionItem": {
            "subjectKind": [
                129
            ],
            "subjectId": [
                2
            ],
            "record": [
                134
            ],
            "position": [
                3
            ],
            "addedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FieldType": {},
        "FieldDef": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "kinds": [
                129
            ],
            "fieldType": [
                44
            ],
            "options": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FieldDefInput": {
            "name": [
                1
            ],
            "kinds": [
                129
            ],
            "fieldType": [
                44
            ],
            "options": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Gallery": {
            "categories": [
                39
            ],
            "related": [
                48,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "files": [
                59
            ],
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                77
            ],
            "status": [
                131
            ],
            "studio": [
                83
            ],
            "performers": [
                81
            ],
            "tags": [
                84
            ],
            "imageCount": [
                3
            ],
            "images": [
                80
            ],
            "video": [
                38
            ],
            "sources": [
                67
            ],
            "userState": [
                135
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                72
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "Match": {
            "item": [
                133
            ],
            "record": [
                134
            ],
            "score": [
                5
            ],
            "evidence": [
                49
            ],
            "__typename": [
                1
            ]
        },
        "Evidence": {
            "plugin": [
                1
            ],
            "url": [
                1
            ],
            "title": [
                1
            ],
            "image": [
                77
            ],
            "detail": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Short": {
            "match": [
                48
            ],
            "video": [
                38
            ],
            "startSeconds": [
                5
            ],
            "endSeconds": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "Taste": {
            "signalCount": [
                3
            ],
            "tags": [
                52
            ],
            "performers": [
                52
            ],
            "studios": [
                52
            ],
            "sites": [
                52
            ],
            "__typename": [
                1
            ]
        },
        "TasteEntry": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "image": [
                77
            ],
            "weight": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "LookupInput": {
            "text": [
                1
            ],
            "imageBase64": [
                1
            ],
            "mimeType": [
                1
            ],
            "plugins": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "LookupResult": {
            "matches": [
                48
            ],
            "failures": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "AlikePair": {
            "id": [
                2
            ],
            "subjectKind": [
                129
            ],
            "first": [
                134
            ],
            "second": [
                134
            ],
            "score": [
                5
            ],
            "reason": [
                1
            ],
            "provablySame": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "LibraryRoot": {
            "id": [
                2
            ],
            "location": [
                127
            ],
            "path": [
                1
            ],
            "scanIntervalMinutes": [
                3
            ],
            "lastScanAt": [
                1
            ],
            "organizeTemplate": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FileMove": {
            "fileId": [
                2
            ],
            "from": [
                1
            ],
            "to": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FileMoveInput": {
            "fileId": [
                2
            ],
            "from": [
                1
            ],
            "to": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "MediaFile": {
            "id": [
                2
            ],
            "subjectKind": [
                129
            ],
            "subjectId": [
                2
            ],
            "url": [
                1
            ],
            "path": [
                1
            ],
            "location": [
                127
            ],
            "sizeBytes": [
                3
            ],
            "mimeType": [
                1
            ],
            "quality": [
                1
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "durationSeconds": [
                5
            ],
            "missing": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "Download": {
            "job": [
                123
            ],
            "file": [
                59
            ],
            "__typename": [
                1
            ]
        },
        "FingerprintAlgorithm": {},
        "Fingerprint": {
            "id": [
                2
            ],
            "algorithm": [
                61
            ],
            "value": [
                1
            ],
            "fileId": [
                2
            ],
            "durationSeconds": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "EmbeddingKind": {},
        "Embedding": {
            "id": [
                2
            ],
            "kind": [
                63
            ],
            "model": [
                1
            ],
            "atSeconds": [
                5
            ],
            "region": [
                128
            ],
            "__typename": [
                1
            ]
        },
        "EmbeddingInput": {
            "subject": [
                132
            ],
            "kind": [
                63
            ],
            "model": [
                1
            ],
            "vector": [
                5
            ],
            "atSeconds": [
                5
            ],
            "region": [
                128
            ],
            "__typename": [
                1
            ]
        },
        "Audio": {
            "files": [
                59
            ],
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                77
            ],
            "status": [
                131
            ],
            "studio": [
                83
            ],
            "performers": [
                81
            ],
            "tags": [
                84
            ],
            "durationSeconds": [
                5
            ],
            "sources": [
                67
            ],
            "streams": [
                85
            ],
            "userState": [
                135
            ],
            "fieldSources": [
                72
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "segments": [
                105
            ],
            "subtitles": [
                110
            ],
            "scripts": [
                113
            ],
            "comments": [
                115
            ],
            "__typename": [
                1
            ]
        },
        "Source": {
            "snapshot": [
                77
            ],
            "id": [
                2
            ],
            "plugin": [
                1
            ],
            "externalId": [
                1
            ],
            "url": [
                1
            ],
            "title": [
                1
            ],
            "image": [
                77
            ],
            "fetchedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "MediaEdit": {
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "studioId": [
                2
            ],
            "performerIds": [
                2
            ],
            "tagIds": [
                2
            ],
            "status": [
                131
            ],
            "vr": [
                86
            ],
            "clearVr": [
                4
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "durationSeconds": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "PerformerEdit": {
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "details": [
                1
            ],
            "links": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "StudioEdit": {
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parentId": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "TagEdit": {
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FieldSource": {
            "field": [
                1
            ],
            "writtenBy": [
                74
            ],
            "source": [
                67
            ],
            "locked": [
                4
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Revision": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "subjectKind": [
                129
            ],
            "subjectId": [
                2
            ],
            "field": [
                1
            ],
            "before": [
                128
            ],
            "after": [
                128
            ],
            "writtenBy": [
                74
            ],
            "actorName": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Actor": {},
        "ListKind": {},
        "Listed": {
            "videos": [
                38
            ],
            "galleries": [
                47
            ],
            "performers": [
                81
            ],
            "studios": [
                83
            ],
            "tags": [
                84
            ],
            "__typename": [
                1
            ]
        },
        "Asset": {
            "uri": [
                1
            ],
            "url": [
                1
            ],
            "mimeType": [
                1
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "blurhash": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "VrFormat": {
            "projection": [
                87
            ],
            "stereo": [
                88
            ],
            "fieldOfView": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "Trickplay": {
            "sheets": [
                77
            ],
            "columns": [
                3
            ],
            "rows": [
                3
            ],
            "tileWidth": [
                3
            ],
            "tileHeight": [
                3
            ],
            "intervalSeconds": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "Image": {
            "id": [
                2
            ],
            "position": [
                3
            ],
            "asset": [
                77
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "Performer": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "image": [
                77
            ],
            "details": [
                1
            ],
            "attributes": [
                82
            ],
            "links": [
                1
            ],
            "tags": [
                84
            ],
            "sources": [
                67
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                135
            ],
            "fieldSources": [
                72
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "PerformerAttributes": {
            "gender": [
                1
            ],
            "birthdate": [
                1
            ],
            "deathDate": [
                1
            ],
            "country": [
                1
            ],
            "ethnicity": [
                1
            ],
            "eyeColor": [
                1
            ],
            "hairColor": [
                1
            ],
            "heightCm": [
                3
            ],
            "weightKg": [
                3
            ],
            "measurements": [
                1
            ],
            "fakeTits": [
                1
            ],
            "tattoos": [
                1
            ],
            "piercings": [
                1
            ],
            "careerStart": [
                3
            ],
            "careerEnd": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "Studio": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "image": [
                77
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                83
            ],
            "tags": [
                84
            ],
            "sources": [
                67
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                135
            ],
            "fieldSources": [
                72
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "Tag": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                135
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "Stream": {
            "id": [
                2
            ],
            "url": [
                1
            ],
            "mimeType": [
                1
            ],
            "quality": [
                1
            ],
            "height": [
                3
            ],
            "speedKbps": [
                3
            ],
            "source": [
                67
            ],
            "expiresAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "VrFormatInput": {
            "projection": [
                87
            ],
            "stereo": [
                88
            ],
            "fieldOfView": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "VrProjection": {},
        "VrStereo": {},
        "Story": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                77
            ],
            "status": [
                131
            ],
            "studio": [
                83
            ],
            "performers": [
                81
            ],
            "tags": [
                84
            ],
            "chapterCount": [
                3
            ],
            "wordCount": [
                3
            ],
            "complete": [
                4
            ],
            "chapters": [
                90
            ],
            "sources": [
                67
            ],
            "userState": [
                135
            ],
            "fieldSources": [
                72
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "StoryChapter": {
            "id": [
                2
            ],
            "position": [
                3
            ],
            "title": [
                1
            ],
            "text": [
                1
            ],
            "wordCount": [
                3
            ],
            "url": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Post": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "title": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                77
            ],
            "status": [
                131
            ],
            "creator": [
                81
            ],
            "text": [
                1
            ],
            "tags": [
                84
            ],
            "attachments": [
                134
            ],
            "locked": [
                4
            ],
            "priceCents": [
                3
            ],
            "currency": [
                1
            ],
            "sources": [
                67
            ],
            "userState": [
                135
            ],
            "fieldSources": [
                72
            ],
            "revisions": [
                73,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "Group": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "title": [
                1
            ],
            "details": [
                1
            ],
            "date": [
                1
            ],
            "cover": [
                77
            ],
            "studio": [
                83
            ],
            "videos": [
                38
            ],
            "sources": [
                67
            ],
            "__typename": [
                1
            ]
        },
        "ApiToken": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "scopes": [
                1
            ],
            "lastUsedAt": [
                1
            ],
            "expiresAt": [
                1
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "CreatedApiToken": {
            "token": [
                1
            ],
            "apiToken": [
                93
            ],
            "__typename": [
                1
            ]
        },
        "AccountStatus": {},
        "Account": {
            "id": [
                2
            ],
            "plugin": [
                1
            ],
            "label": [
                1
            ],
            "status": [
                95
            ],
            "expiresAt": [
                1
            ],
            "lastCheckedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginContribution": {
            "plugin": [
                1
            ],
            "kind": [
                1
            ],
            "name": [
                1
            ],
            "title": [
                1
            ],
            "mediaKinds": [
                1
            ],
            "domains": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginView": {
            "kind": [
                1
            ],
            "title": [
                1
            ],
            "body": [
                1
            ],
            "fields": [
                99
            ],
            "items": [
                100
            ],
            "actions": [
                101
            ],
            "__typename": [
                1
            ]
        },
        "PluginViewField": {
            "key": [
                1
            ],
            "label": [
                1
            ],
            "type": [
                1
            ],
            "value": [
                1
            ],
            "options": [
                1
            ],
            "required": [
                4
            ],
            "hint": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginViewItem": {
            "title": [
                1
            ],
            "subtitle": [
                1
            ],
            "imageUrl": [
                1
            ],
            "url": [
                1
            ],
            "record": [
                134
            ],
            "__typename": [
                1
            ]
        },
        "PluginViewAction": {
            "name": [
                1
            ],
            "label": [
                1
            ],
            "primary": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "PluginViewResult": {
            "view": [
                98
            ],
            "message": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Plugin": {
            "contributions": [
                97
            ],
            "account": [
                96
            ],
            "name": [
                1
            ],
            "packageName": [
                1
            ],
            "localBuild": [
                4
            ],
            "displayName": [
                1
            ],
            "iconUrl": [
                1
            ],
            "description": [
                1
            ],
            "version": [
                1
            ],
            "capabilities": [
                1
            ],
            "domains": [
                1
            ],
            "content": [
                1
            ],
            "enabled": [
                4
            ],
            "requiresSolver": [
                4
            ],
            "available": [
                4
            ],
            "settings": [
                117
            ],
            "settingValues": [
                118
            ],
            "__typename": [
                1
            ]
        },
        "SegmentOrigin": {},
        "Segment": {
            "id": [
                2
            ],
            "subjectKind": [
                129
            ],
            "subjectId": [
                2
            ],
            "startSeconds": [
                5
            ],
            "endSeconds": [
                5
            ],
            "title": [
                1
            ],
            "tag": [
                84
            ],
            "performers": [
                81
            ],
            "origin": [
                104
            ],
            "__typename": [
                1
            ]
        },
        "SegmentInput": {
            "startSeconds": [
                5
            ],
            "endSeconds": [
                5
            ],
            "clearEnd": [
                4
            ],
            "title": [
                1
            ],
            "tagId": [
                2
            ],
            "performerIds": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Clip": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "video": [
                38
            ],
            "startSeconds": [
                5
            ],
            "endSeconds": [
                5
            ],
            "title": [
                1
            ],
            "cover": [
                77
            ],
            "file": [
                59
            ],
            "__typename": [
                1
            ]
        },
        "ClipInput": {
            "startSeconds": [
                5
            ],
            "endSeconds": [
                5
            ],
            "title": [
                1
            ],
            "coverUri": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SubtitleOrigin": {},
        "Subtitle": {
            "id": [
                2
            ],
            "language": [
                1
            ],
            "format": [
                1
            ],
            "asset": [
                77
            ],
            "origin": [
                109
            ],
            "__typename": [
                1
            ]
        },
        "SubtitleInput": {
            "language": [
                1
            ],
            "format": [
                1
            ],
            "uri": [
                1
            ],
            "origin": [
                109
            ],
            "__typename": [
                1
            ]
        },
        "ScriptKind": {},
        "Script": {
            "id": [
                2
            ],
            "kind": [
                112
            ],
            "asset": [
                77
            ],
            "axis": [
                1
            ],
            "writtenBy": [
                74
            ],
            "__typename": [
                1
            ]
        },
        "TrickplayInput": {
            "sheetUris": [
                1
            ],
            "columns": [
                3
            ],
            "rows": [
                3
            ],
            "tileWidth": [
                3
            ],
            "tileHeight": [
                3
            ],
            "intervalSeconds": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "Comment": {
            "id": [
                2
            ],
            "source": [
                67
            ],
            "author": [
                1
            ],
            "text": [
                1
            ],
            "postedAt": [
                1
            ],
            "likes": [
                3
            ],
            "replyTo": [
                2
            ],
            "mentions": [
                81
            ],
            "__typename": [
                1
            ]
        },
        "PluginPackage": {
            "name": [
                1
            ],
            "version": [
                1
            ],
            "description": [
                1
            ],
            "installedVersion": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginSettingField": {
            "key": [
                1
            ],
            "label": [
                1
            ],
            "description": [
                1
            ],
            "type": [
                1
            ],
            "required": [
                4
            ],
            "default": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginSettingValue": {
            "key": [
                1
            ],
            "value": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginSettingValueInput": {
            "key": [
                1
            ],
            "value": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "TrashedRecord": {
            "kind": [
                129
            ],
            "id": [
                2
            ],
            "label": [
                1
            ],
            "deletedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "JobKind": {},
        "JobStatus": {},
        "Job": {
            "id": [
                2
            ],
            "kind": [
                121
            ],
            "status": [
                122
            ],
            "subject": [
                133
            ],
            "plugin": [
                1
            ],
            "payload": [
                128
            ],
            "progress": [
                5
            ],
            "error": [
                1
            ],
            "attempts": [
                3
            ],
            "priority": [
                3
            ],
            "device": [
                2
            ],
            "createdAt": [
                1
            ],
            "runAt": [
                1
            ],
            "startedAt": [
                1
            ],
            "finishedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Platform": {},
        "Device": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "platform": [
                124
            ],
            "lastSeenAt": [
                1
            ],
            "lastSyncAt": [
                1
            ],
            "current": [
                4
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "StorageKind": {},
        "StorageLocation": {
            "id": [
                2
            ],
            "device": [
                125
            ],
            "kind": [
                126
            ],
            "uri": [
                1
            ],
            "quotaBytes": [
                3
            ],
            "usedBytes": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "JSON": {},
        "RecordKind": {},
        "MediaKind": {},
        "MediaStatus": {},
        "RecordRefInput": {
            "kind": [
                129
            ],
            "id": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "RecordRef": {
            "kind": [
                129
            ],
            "id": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Record": {
            "on_Video": [
                38
            ],
            "on_Gallery": [
                47
            ],
            "on_Audio": [
                66
            ],
            "on_Story": [
                89
            ],
            "on_Post": [
                91
            ],
            "on_Performer": [
                81
            ],
            "on_Studio": [
                83
            ],
            "on_Tag": [
                84
            ],
            "on_Group": [
                92
            ],
            "on_Category": [
                39
            ],
            "on_Collection": [
                42
            ],
            "on_CamModel": [
                35
            ],
            "on_Segment": [
                105
            ],
            "on_Clip": [
                107
            ],
            "__typename": [
                1
            ]
        },
        "UserState": {
            "subjectKind": [
                129
            ],
            "subjectId": [
                2
            ],
            "record": [
                134
            ],
            "rating": [
                5
            ],
            "favorite": [
                4
            ],
            "notes": [
                1
            ],
            "blocked": [
                4
            ],
            "progress": [
                5
            ],
            "viewCount": [
                3
            ],
            "lastViewedAt": [
                1
            ],
            "fields": [
                128
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "UserStateFilter": {
            "kind": [
                129
            ],
            "favorite": [
                4
            ],
            "blocked": [
                4
            ],
            "rated": [
                4
            ],
            "viewed": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "EventKind": {},
        "Event": {
            "id": [
                2
            ],
            "at": [
                1
            ],
            "kind": [
                137
            ],
            "subjectKind": [
                129
            ],
            "subjectId": [
                2
            ],
            "data": [
                128
            ],
            "device": [
                2
            ],
            "sessionId": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "EventInput": {
            "kind": [
                137
            ],
            "subject": [
                132
            ],
            "data": [
                128
            ],
            "sessionId": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "EventFilter": {
            "kinds": [
                137
            ],
            "subject": [
                132
            ],
            "since": [
                1
            ],
            "until": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SortField": {},
        "Sort": {
            "field": [
                141
            ],
            "descending": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "SortInput": {
            "field": [
                141
            ],
            "descending": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "FieldCondition": {
            "fieldDefId": [
                2
            ],
            "operator": [
                1
            ],
            "value": [
                128
            ],
            "__typename": [
                1
            ]
        },
        "FieldConditionInput": {
            "fieldDefId": [
                2
            ],
            "operator": [
                1
            ],
            "value": [
                128
            ],
            "__typename": [
                1
            ]
        },
        "Filter": {
            "kinds": [
                130
            ],
            "text": [
                1
            ],
            "performerIds": [
                2
            ],
            "studioIds": [
                2
            ],
            "categoryIds": [
                2
            ],
            "includeTagIds": [
                2
            ],
            "excludeTagIds": [
                2
            ],
            "sources": [
                1
            ],
            "dateFrom": [
                1
            ],
            "dateTo": [
                1
            ],
            "minDurationSeconds": [
                3
            ],
            "maxDurationSeconds": [
                3
            ],
            "minRating": [
                5
            ],
            "status": [
                131
            ],
            "inLibrary": [
                4
            ],
            "favorite": [
                4
            ],
            "unwatched": [
                4
            ],
            "vr": [
                4
            ],
            "fields": [
                144
            ],
            "sort": [
                142
            ],
            "__typename": [
                1
            ]
        },
        "FilterInput": {
            "kinds": [
                130
            ],
            "text": [
                1
            ],
            "performerIds": [
                2
            ],
            "studioIds": [
                2
            ],
            "categoryIds": [
                2
            ],
            "includeTagIds": [
                2
            ],
            "excludeTagIds": [
                2
            ],
            "sources": [
                1
            ],
            "dateFrom": [
                1
            ],
            "dateTo": [
                1
            ],
            "minDurationSeconds": [
                3
            ],
            "maxDurationSeconds": [
                3
            ],
            "minRating": [
                5
            ],
            "status": [
                131
            ],
            "inLibrary": [
                4
            ],
            "favorite": [
                4
            ],
            "unwatched": [
                4
            ],
            "vr": [
                4
            ],
            "fields": [
                145
            ],
            "sort": [
                143
            ],
            "__typename": [
                1
            ]
        },
        "SavedFilter": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "filter": [
                146
            ],
            "pinned": [
                4
            ],
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        }
    }
}