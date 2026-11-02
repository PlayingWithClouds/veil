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
        70,
        72,
        85,
        86,
        102,
        103,
        110,
        119,
        124,
        127,
        136,
        137,
        139,
        141,
        143,
        144,
        145,
        146,
        152,
        157
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
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "newSince": [
                149,
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
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "fieldDefs": [
                45,
                {
                    "kind": [
                        144
                    ]
                }
            ],
            "faceScans": [
                48,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "facePictures": [
                51,
                {
                    "performerId": [
                        2,
                        "ID!"
                    ],
                    "model": [
                        1,
                        "String!"
                    ]
                }
            ],
            "videoFaces": [
                52,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "model": [
                        1
                    ]
                }
            ],
            "recommendations": [
                57,
                {
                    "filter": [
                        163
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "related": [
                57,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "shorts": [
                59,
                {
                    "filter": [
                        163
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "taste": [
                60
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
                64,
                {
                    "kind": [
                        144,
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
                65
            ],
            "libraryRoot": [
                65,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "planOrganize": [
                66,
                {
                    "rootId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "downloads": [
                69
            ],
            "sameMedia": [
                149,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "nearest": [
                57,
                {
                    "kind": [
                        72,
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
                        144,
                        "[RecordKind!]"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "similar": [
                57,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "kind": [
                        72,
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
                        163
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
                        163
                    ]
                }
            ],
            "videoTagCounts": [
                87,
                {
                    "filter": [
                        163
                    ],
                    "limit": [
                        3
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
                        163
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
                        163
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
                93,
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
                93,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "performerDuplicates": [
                93,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "collecting": [
                95
            ],
            "performerAttributeValues": [
                1,
                {
                    "attribute": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "studios": [
                98,
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
                98,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "tags": [
                99,
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
                99,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "media": [
                149,
                {
                    "filter": [
                        163
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
                        163
                    ]
                }
            ],
            "audios": [
                75,
                {
                    "filter": [
                        163
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
                        163
                    ]
                }
            ],
            "audio": [
                75,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "stories": [
                104,
                {
                    "filter": [
                        163
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
                        163
                    ]
                }
            ],
            "story": [
                104,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "posts": [
                106,
                {
                    "filter": [
                        163
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
                        163
                    ]
                }
            ],
            "post": [
                106,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "groups": [
                107,
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
                107,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "settings": [
                143
            ],
            "apiTokens": [
                108
            ],
            "accounts": [
                111
            ],
            "contributions": [
                112,
                {
                    "kind": [
                        1
                    ]
                }
            ],
            "pluginView": [
                113,
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
                        143
                    ]
                }
            ],
            "clips": [
                122,
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
                122,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                118
            ],
            "plugin": [
                118,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "pluginPackages": [
                131,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "trash": [
                135,
                {
                    "kind": [
                        144,
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
                138,
                {
                    "statuses": [
                        137,
                        "[JobStatus!]"
                    ],
                    "kinds": [
                        136,
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
                138,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "devices": [
                140
            ],
            "currentDevice": [
                140
            ],
            "storageLocations": [
                142,
                {
                    "deviceId": [
                        2
                    ]
                }
            ],
            "userStates": [
                150,
                {
                    "filter": [
                        151
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
                150,
                {
                    "kind": [
                        144
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
                154,
                {
                    "filter": [
                        156
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
                        156
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
            "sitePreferences": [
                153
            ],
            "savedFilters": [
                164
            ],
            "savedFilter": [
                164,
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
                        147,
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
                138,
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
            "clearNotifications": [
                4,
                {
                    "ids": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "clearAllNotifications": [
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
                        143
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
                        143
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
                        147,
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
                138,
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
                        147,
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
                        147,
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
                        147,
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
                150,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "fieldDefId": [
                        2,
                        "ID!"
                    ],
                    "value": [
                        143
                    ]
                }
            ],
            "recordFaceScan": [
                48,
                {
                    "input": [
                        49,
                        "FaceScanInput!"
                    ]
                }
            ],
            "learnFaces": [
                3,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "model": [
                        1
                    ]
                }
            ],
            "lookup": [
                63,
                {
                    "input": [
                        62,
                        "LookupInput!"
                    ]
                }
            ],
            "merge": [
                149,
                {
                    "kind": [
                        144,
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
                149,
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
                65,
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
                65,
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
                138,
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
                        67,
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
                138,
                {
                    "subject": [
                        147,
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
                71,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "fileId": [
                        2
                    ],
                    "algorithm": [
                        70,
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
                138,
                {
                    "fileId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "snapshotSource": [
                138,
                {
                    "sourceId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addEmbedding": [
                73,
                {
                    "input": [
                        74,
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
            "fetchPerformer": [
                93,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "refresh": [
                        4
                    ]
                }
            ],
            "unassignPerformer": [
                93,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "performerId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "browse": [
                88,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "kind": [
                        86,
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
                88,
                {
                    "kind": [
                        86,
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
                        78,
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
                        78,
                        "MediaEdit!"
                    ]
                }
            ],
            "editPerformer": [
                93,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        79,
                        "PerformerEdit!"
                    ]
                }
            ],
            "editStudio": [
                98,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        81,
                        "StudioEdit!"
                    ]
                }
            ],
            "editTag": [
                99,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        82,
                        "TagEdit!"
                    ]
                }
            ],
            "lockField": [
                83,
                {
                    "subject": [
                        147,
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
                149,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "edit": [
                        78,
                        "MediaEdit!"
                    ]
                }
            ],
            "fetchMedia": [
                149,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "readChapter": [
                105,
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
                143,
                {
                    "values": [
                        143,
                        "JSON!"
                    ]
                }
            ],
            "createApiToken": [
                109,
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
                111,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "values": [
                        143,
                        "JSON!"
                    ]
                }
            ],
            "checkAccount": [
                111,
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
                117,
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
                        143
                    ],
                    "input": [
                        143
                    ]
                }
            ],
            "addSegment": [
                120,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "input": [
                        121,
                        "SegmentInput!"
                    ]
                }
            ],
            "editSegment": [
                120,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        121,
                        "SegmentInput!"
                    ]
                }
            ],
            "addClip": [
                122,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        123,
                        "ClipInput!"
                    ]
                }
            ],
            "editClip": [
                122,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        123,
                        "ClipInput!"
                    ]
                }
            ],
            "attachClipFile": [
                122,
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
                125,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "input": [
                        126,
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
                128,
                {
                    "subject": [
                        147,
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
                        129
                    ]
                }
            ],
            "reportStreamSpeed": [
                100,
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
                149,
                {
                    "subject": [
                        147,
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
                        134,
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
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "restoreRecord": [
                4,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "purgeRecord": [
                4,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "emptyTrash": [
                3
            ],
            "queueFetch": [
                138,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ]
                }
            ],
            "cancelJob": [
                138,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "retryJob": [
                138,
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
                140,
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
                142,
                {
                    "kind": [
                        141,
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
                142,
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
                150,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "rating": [
                        5
                    ]
                }
            ],
            "setFavorite": [
                150,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "favorite": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setBlocked": [
                150,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "blocked": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setNotes": [
                150,
                {
                    "subject": [
                        147,
                        "RecordRefInput!"
                    ],
                    "notes": [
                        1,
                        "String!"
                    ]
                }
            ],
            "recordEvent": [
                154,
                {
                    "event": [
                        155,
                        "EventInput!"
                    ]
                }
            ],
            "recordEvents": [
                3,
                {
                    "events": [
                        155,
                        "[EventInput!]!"
                    ]
                }
            ],
            "clearEvents": [
                3,
                {
                    "kinds": [
                        152,
                        "[EventKind!]"
                    ]
                }
            ],
            "saveFilter": [
                164,
                {
                    "id": [
                        2
                    ],
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        163,
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
            "searchResults": [
                77,
                {
                    "kind": [
                        86,
                        "ListKind!"
                    ],
                    "query": [
                        1,
                        "String!"
                    ],
                    "page": [
                        3
                    ],
                    "limit": [
                        3
                    ],
                    "plugins": [
                        1,
                        "[String!]"
                    ],
                    "filter": [
                        163
                    ]
                }
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
                162
            ],
            "targetKind": [
                144
            ],
            "targetId": [
                2
            ],
            "target": [
                149
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
                163
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
                163
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
                148
            ],
            "record": [
                149
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
                143
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
                162
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
                163
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
                143
            ],
            "condition": [
                162
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
                143
            ],
            "condition": [
                163
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
                143
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
                143
            ],
            "error": [
                1
            ],
            "touched": [
                148
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
                93
            ],
            "thumbnail": [
                89
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
                150
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
                89
            ],
            "__typename": [
                1
            ]
        },
        "ActiveRecording": {
            "job": [
                138
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
                57,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "files": [
                68
            ],
            "fingerprints": [
                71
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
                89
            ],
            "status": [
                146
            ],
            "studio": [
                98
            ],
            "performers": [
                93
            ],
            "suggestedPerformers": [
                94
            ],
            "tags": [
                99
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
                90
            ],
            "preview": [
                89
            ],
            "trickplay": [
                91
            ],
            "previewFrames": [
                89
            ],
            "gallery": [
                47
            ],
            "sources": [
                76
            ],
            "streams": [
                100
            ],
            "userState": [
                150
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                83
            ],
            "revisions": [
                84,
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
                107
            ],
            "segments": [
                120
            ],
            "subtitles": [
                125
            ],
            "scripts": [
                128
            ],
            "clips": [
                122
            ],
            "comments": [
                130
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
                99
            ],
            "cover": [
                89
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
                89
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
                144
            ],
            "subjectId": [
                2
            ],
            "record": [
                149
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
                144
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
                144
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
                57,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "files": [
                68
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
                89
            ],
            "status": [
                146
            ],
            "studio": [
                98
            ],
            "performers": [
                93
            ],
            "tags": [
                99
            ],
            "imageCount": [
                3
            ],
            "images": [
                92
            ],
            "video": [
                38
            ],
            "sources": [
                76
            ],
            "userState": [
                150
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                83
            ],
            "revisions": [
                84,
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
        "FaceScan": {
            "id": [
                2
            ],
            "subjectKind": [
                144
            ],
            "subjectId": [
                2
            ],
            "model": [
                1
            ],
            "picture": [
                1
            ],
            "frames": [
                3
            ],
            "faces": [
                3
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FaceScanInput": {
            "subject": [
                147
            ],
            "model": [
                1
            ],
            "picture": [
                1
            ],
            "frames": [
                3
            ],
            "faces": [
                50
            ],
            "__typename": [
                1
            ]
        },
        "FaceInput": {
            "vector": [
                5
            ],
            "atSeconds": [
                5
            ],
            "region": [
                143
            ],
            "__typename": [
                1
            ]
        },
        "FacePicture": {
            "uri": [
                1
            ],
            "scanned": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "VideoFaces": {
            "scan": [
                48
            ],
            "people": [
                53
            ],
            "guesses": [
                56
            ],
            "__typename": [
                1
            ]
        },
        "FacePerson": {
            "sightings": [
                54
            ],
            "candidates": [
                55
            ],
            "__typename": [
                1
            ]
        },
        "FaceSighting": {
            "embeddingId": [
                2
            ],
            "atSeconds": [
                5
            ],
            "region": [
                143
            ],
            "__typename": [
                1
            ]
        },
        "FaceCandidate": {
            "performer": [
                93
            ],
            "similarity": [
                5
            ],
            "likely": [
                4
            ],
            "credited": [
                4
            ],
            "picture": [
                1
            ],
            "video": [
                38
            ],
            "__typename": [
                1
            ]
        },
        "FaceGuess": {
            "performer": [
                93
            ],
            "similarity": [
                5
            ],
            "credited": [
                4
            ],
            "person": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "Match": {
            "item": [
                148
            ],
            "record": [
                149
            ],
            "score": [
                5
            ],
            "evidence": [
                58
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
                89
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
                57
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
                61
            ],
            "performers": [
                61
            ],
            "studios": [
                61
            ],
            "sites": [
                61
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
                89
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
                57
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
                144
            ],
            "first": [
                149
            ],
            "second": [
                149
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
                142
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
                144
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
                142
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
                138
            ],
            "file": [
                68
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
                70
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
                72
            ],
            "model": [
                1
            ],
            "atSeconds": [
                5
            ],
            "region": [
                143
            ],
            "__typename": [
                1
            ]
        },
        "EmbeddingInput": {
            "subject": [
                147
            ],
            "kind": [
                72
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
                143
            ],
            "__typename": [
                1
            ]
        },
        "Audio": {
            "files": [
                68
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
                89
            ],
            "status": [
                146
            ],
            "studio": [
                98
            ],
            "performers": [
                93
            ],
            "tags": [
                99
            ],
            "durationSeconds": [
                5
            ],
            "sources": [
                76
            ],
            "streams": [
                100
            ],
            "userState": [
                150
            ],
            "fieldSources": [
                83
            ],
            "revisions": [
                84,
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
                120
            ],
            "subtitles": [
                125
            ],
            "scripts": [
                128
            ],
            "comments": [
                130
            ],
            "__typename": [
                1
            ]
        },
        "Source": {
            "snapshot": [
                89
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
                89
            ],
            "fetchedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SearchBatch": {
            "plugin": [
                1
            ],
            "listed": [
                88
            ],
            "sites": [
                3
            ],
            "stored": [
                4
            ],
            "final": [
                4
            ],
            "exhausted": [
                4
            ],
            "returned": [
                3
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
                146
            ],
            "vr": [
                101
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
            "attributes": [
                80
            ],
            "imageUrl": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PerformerAttributesInput": {
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
                85
            ],
            "source": [
                76
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
                144
            ],
            "subjectId": [
                2
            ],
            "field": [
                1
            ],
            "before": [
                143
            ],
            "after": [
                143
            ],
            "writtenBy": [
                85
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
        "TagCount": {
            "tag": [
                99
            ],
            "count": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "Listed": {
            "videos": [
                38
            ],
            "galleries": [
                47
            ],
            "performers": [
                93
            ],
            "studios": [
                98
            ],
            "tags": [
                99
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
                102
            ],
            "stereo": [
                103
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
                89
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
                89
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
                89
            ],
            "details": [
                1
            ],
            "attributes": [
                97
            ],
            "links": [
                1
            ],
            "tags": [
                99
            ],
            "sources": [
                76
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                150
            ],
            "fieldSources": [
                83
            ],
            "revisions": [
                84,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "fetching": [
                96
            ],
            "__typename": [
                1
            ]
        },
        "PerformerSuggestion": {
            "performer": [
                93
            ],
            "reason": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Collecting": {
            "performer": [
                93
            ],
            "reading": [
                4
            ],
            "listing": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PerformerFetching": {
            "reading": [
                4
            ],
            "listing": [
                1
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
                89
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                98
            ],
            "tags": [
                99
            ],
            "sources": [
                76
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                150
            ],
            "fieldSources": [
                83
            ],
            "revisions": [
                84,
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
                150
            ],
            "revisions": [
                84,
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
                76
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
                102
            ],
            "stereo": [
                103
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
                89
            ],
            "status": [
                146
            ],
            "studio": [
                98
            ],
            "performers": [
                93
            ],
            "tags": [
                99
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
                105
            ],
            "sources": [
                76
            ],
            "userState": [
                150
            ],
            "fieldSources": [
                83
            ],
            "revisions": [
                84,
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
                89
            ],
            "status": [
                146
            ],
            "creator": [
                93
            ],
            "text": [
                1
            ],
            "tags": [
                99
            ],
            "attachments": [
                149
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
                76
            ],
            "userState": [
                150
            ],
            "fieldSources": [
                83
            ],
            "revisions": [
                84,
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
                89
            ],
            "studio": [
                98
            ],
            "videos": [
                38
            ],
            "sources": [
                76
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
                108
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
                110
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
                114
            ],
            "items": [
                115
            ],
            "actions": [
                116
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
                149
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
                113
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
                112
            ],
            "account": [
                111
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
                132
            ],
            "settingValues": [
                133
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
                144
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
                99
            ],
            "performers": [
                93
            ],
            "origin": [
                119
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
                89
            ],
            "file": [
                68
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
                89
            ],
            "origin": [
                124
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
                124
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
                127
            ],
            "asset": [
                89
            ],
            "axis": [
                1
            ],
            "writtenBy": [
                85
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
                76
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
                93
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
                144
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
                136
            ],
            "status": [
                137
            ],
            "subject": [
                148
            ],
            "plugin": [
                1
            ],
            "payload": [
                143
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
                139
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
                140
            ],
            "kind": [
                141
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
                144
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
                144
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
                75
            ],
            "on_Story": [
                104
            ],
            "on_Post": [
                106
            ],
            "on_Performer": [
                93
            ],
            "on_Studio": [
                98
            ],
            "on_Tag": [
                99
            ],
            "on_Group": [
                107
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
                120
            ],
            "on_Clip": [
                122
            ],
            "__typename": [
                1
            ]
        },
        "UserState": {
            "subjectKind": [
                144
            ],
            "subjectId": [
                2
            ],
            "record": [
                149
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
                143
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
                144
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
        "SitePreference": {
            "plugin": [
                1
            ],
            "shown": [
                3
            ],
            "chosen": [
                3
            ],
            "score": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "Event": {
            "id": [
                2
            ],
            "at": [
                1
            ],
            "kind": [
                152
            ],
            "subjectKind": [
                144
            ],
            "subjectId": [
                2
            ],
            "data": [
                143
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
                152
            ],
            "subject": [
                147
            ],
            "data": [
                143
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
                152
            ],
            "subject": [
                147
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
                157
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
                157
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
                143
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
                143
            ],
            "__typename": [
                1
            ]
        },
        "Filter": {
            "kinds": [
                145
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
                146
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
            "minHeight": [
                3
            ],
            "fields": [
                160
            ],
            "sort": [
                158
            ],
            "__typename": [
                1
            ]
        },
        "FilterInput": {
            "kinds": [
                145
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
                146
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
            "minHeight": [
                3
            ],
            "fields": [
                161
            ],
            "sort": [
                159
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
                162
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