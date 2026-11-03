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
        42,
        45,
        73,
        75,
        88,
        89,
        105,
        106,
        113,
        122,
        127,
        130,
        140,
        141,
        143,
        145,
        147,
        148,
        149,
        150,
        156,
        161
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
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "newSince": [
                153,
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
            "categoryTiles": [
                40,
                {
                    "kind": [
                        149,
                        "MediaKind!"
                    ]
                }
            ],
            "collections": [
                43
            ],
            "collection": [
                43,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "builtInCollection": [
                43,
                {
                    "kind": [
                        42,
                        "CollectionKind!"
                    ]
                }
            ],
            "collectionsWith": [
                43,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "fieldDefs": [
                46,
                {
                    "kind": [
                        148
                    ]
                }
            ],
            "faceScans": [
                49,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "facePictures": [
                52,
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
                53,
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
            "performerFaceMatches": [
                57,
                {
                    "performerId": [
                        2,
                        "ID!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "recommendations": [
                59,
                {
                    "filter": [
                        167
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
                59,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "shorts": [
                61,
                {
                    "filter": [
                        167
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "taste": [
                62
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
                67,
                {
                    "kind": [
                        148,
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
            "merges": [
                66,
                {
                    "kind": [
                        148,
                        "RecordKind!"
                    ],
                    "keptId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "libraryRoots": [
                68
            ],
            "libraryRoot": [
                68,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "planOrganize": [
                69,
                {
                    "rootId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "downloads": [
                72
            ],
            "sameMedia": [
                153,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "nearest": [
                59,
                {
                    "kind": [
                        75,
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
                        148,
                        "[RecordKind!]"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "similar": [
                59,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "kind": [
                        75,
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
                        167
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
                        167
                    ]
                }
            ],
            "videoTagCounts": [
                90,
                {
                    "filter": [
                        167
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
                48,
                {
                    "filter": [
                        167
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
                        167
                    ]
                }
            ],
            "gallery": [
                48,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "performers": [
                96,
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
                96,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "performerDuplicates": [
                96,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "collecting": [
                98
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
                101,
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
                101,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "tags": [
                102,
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
                102,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "media": [
                153,
                {
                    "filter": [
                        167
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
                        167
                    ]
                }
            ],
            "mediaTagCounts": [
                90,
                {
                    "kind": [
                        149,
                        "MediaKind!"
                    ],
                    "filter": [
                        167
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "creditedPerformers": [
                96,
                {
                    "kind": [
                        149,
                        "MediaKind!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "audios": [
                78,
                {
                    "filter": [
                        167
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
                        167
                    ]
                }
            ],
            "audio": [
                78,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "similarAudios": [
                78,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "stories": [
                107,
                {
                    "filter": [
                        167
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
                        167
                    ]
                }
            ],
            "story": [
                107,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "similarStories": [
                107,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "posts": [
                109,
                {
                    "filter": [
                        167
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
                        167
                    ]
                }
            ],
            "post": [
                109,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "groups": [
                110,
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
                110,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "settings": [
                147
            ],
            "apiTokens": [
                111
            ],
            "accounts": [
                114
            ],
            "contributions": [
                115,
                {
                    "kind": [
                        1
                    ]
                }
            ],
            "pluginView": [
                116,
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
                        147
                    ]
                }
            ],
            "clips": [
                125,
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
                125,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                121
            ],
            "plugin": [
                121,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "pluginPackages": [
                134,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "trash": [
                139,
                {
                    "kind": [
                        148,
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
                142,
                {
                    "statuses": [
                        141,
                        "[JobStatus!]"
                    ],
                    "kinds": [
                        140,
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
                142,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "devices": [
                144
            ],
            "currentDevice": [
                144
            ],
            "storageLocations": [
                146,
                {
                    "deviceId": [
                        2
                    ]
                }
            ],
            "userStates": [
                154,
                {
                    "filter": [
                        155
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
                154,
                {
                    "kind": [
                        148
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
                158,
                {
                    "filter": [
                        160
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
                        160
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
                157
            ],
            "savedFilters": [
                168
            ],
            "savedFilter": [
                168,
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
                        151,
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
                142,
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
                        147
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
                        147
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
                        151,
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
                142,
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
                        41,
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
                        41,
                        "CategoryInput!"
                    ]
                }
            ],
            "deleteCategory": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addDefaultCategories": [
                3
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
                43,
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
                43,
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
                43,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subjects": [
                        151,
                        "[RecordRefInput!]!"
                    ],
                    "position": [
                        3
                    ]
                }
            ],
            "removeFromCollection": [
                43,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subjects": [
                        151,
                        "[RecordRefInput!]!"
                    ]
                }
            ],
            "moveInCollection": [
                43,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "position": [
                        3,
                        "Int!"
                    ]
                }
            ],
            "clearCollection": [
                43,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "saveFieldDef": [
                46,
                {
                    "id": [
                        2
                    ],
                    "input": [
                        47,
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
                154,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "fieldDefId": [
                        2,
                        "ID!"
                    ],
                    "value": [
                        147
                    ]
                }
            ],
            "recordFaceScan": [
                49,
                {
                    "input": [
                        50,
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
                65,
                {
                    "input": [
                        64,
                        "LookupInput!"
                    ]
                }
            ],
            "merge": [
                153,
                {
                    "kind": [
                        148,
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
            "undoMerge": [
                153,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "mergeAlike": [
                153,
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
                68,
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
                68,
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
                142,
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
                        70,
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
                142,
                {
                    "subject": [
                        151,
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
                74,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "fileId": [
                        2
                    ],
                    "algorithm": [
                        73,
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
                142,
                {
                    "fileId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "snapshotSource": [
                142,
                {
                    "sourceId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addEmbedding": [
                76,
                {
                    "input": [
                        77,
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
                48,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "fetchPerformer": [
                96,
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
                96,
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
            "createPerformer": [
                96,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "createStudio": [
                101,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "browse": [
                91,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "kind": [
                        89,
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
                91,
                {
                    "kind": [
                        89,
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
                        81,
                        "MediaEdit!"
                    ]
                }
            ],
            "editGallery": [
                48,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        81,
                        "MediaEdit!"
                    ]
                }
            ],
            "editPerformer": [
                96,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        82,
                        "PerformerEdit!"
                    ]
                }
            ],
            "editStudio": [
                101,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        84,
                        "StudioEdit!"
                    ]
                }
            ],
            "editTag": [
                102,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        85,
                        "TagEdit!"
                    ]
                }
            ],
            "lockField": [
                86,
                {
                    "subject": [
                        151,
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
                153,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "edit": [
                        81,
                        "MediaEdit!"
                    ]
                }
            ],
            "fetchMedia": [
                153,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "readChapter": [
                108,
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
                147,
                {
                    "values": [
                        147,
                        "JSON!"
                    ]
                }
            ],
            "createApiToken": [
                112,
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
                114,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "values": [
                        147,
                        "JSON!"
                    ]
                }
            ],
            "importAccountCookies": [
                114,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "cookies": [
                        1,
                        "String!"
                    ]
                }
            ],
            "checkAccount": [
                114,
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
                120,
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
                        147
                    ],
                    "input": [
                        147
                    ]
                }
            ],
            "addSegment": [
                123,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "input": [
                        124,
                        "SegmentInput!"
                    ]
                }
            ],
            "editSegment": [
                123,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        124,
                        "SegmentInput!"
                    ]
                }
            ],
            "addClip": [
                125,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        126,
                        "ClipInput!"
                    ]
                }
            ],
            "editClip": [
                125,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        126,
                        "ClipInput!"
                    ]
                }
            ],
            "attachClipFile": [
                125,
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
                128,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "input": [
                        129,
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
                131,
                {
                    "subject": [
                        151,
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
                        132
                    ]
                }
            ],
            "reportStreamSpeed": [
                103,
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
                153,
                {
                    "subject": [
                        151,
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
                        138,
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
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "restoreRecord": [
                4,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "purgeRecord": [
                4,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "emptyTrash": [
                3
            ],
            "queueFetch": [
                142,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ]
                }
            ],
            "cancelJob": [
                142,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "retryJob": [
                142,
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
                144,
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
                146,
                {
                    "kind": [
                        145,
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
                146,
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
                154,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "rating": [
                        5
                    ]
                }
            ],
            "setFavorite": [
                154,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "favorite": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setBlocked": [
                154,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "blocked": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setNotes": [
                154,
                {
                    "subject": [
                        151,
                        "RecordRefInput!"
                    ],
                    "notes": [
                        1,
                        "String!"
                    ]
                }
            ],
            "recordEvent": [
                158,
                {
                    "event": [
                        159,
                        "EventInput!"
                    ]
                }
            ],
            "recordEvents": [
                3,
                {
                    "events": [
                        159,
                        "[EventInput!]!"
                    ]
                }
            ],
            "clearEvents": [
                3,
                {
                    "kinds": [
                        156,
                        "[EventKind!]"
                    ]
                }
            ],
            "saveFilter": [
                168,
                {
                    "id": [
                        2
                    ],
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        167,
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
                80,
                {
                    "kind": [
                        89,
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
                        167
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
                166
            ],
            "targetKind": [
                148
            ],
            "targetId": [
                2
            ],
            "target": [
                153
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
                167
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
                167
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
                152
            ],
            "record": [
                153
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
                147
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
                166
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
                167
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
                147
            ],
            "condition": [
                166
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
                147
            ],
            "condition": [
                167
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
                147
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
                147
            ],
            "error": [
                1
            ],
            "touched": [
                152
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
                96
            ],
            "thumbnail": [
                92
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
                154
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
                92
            ],
            "__typename": [
                1
            ]
        },
        "ActiveRecording": {
            "job": [
                142
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
                59,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "files": [
                71
            ],
            "fingerprints": [
                74
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
                92
            ],
            "status": [
                150
            ],
            "studio": [
                101
            ],
            "performers": [
                96
            ],
            "suggestedPerformers": [
                97
            ],
            "tags": [
                102
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
                93
            ],
            "preview": [
                92
            ],
            "trickplay": [
                94
            ],
            "previewFrames": [
                92
            ],
            "gallery": [
                48
            ],
            "sources": [
                79
            ],
            "streams": [
                103
            ],
            "userState": [
                154
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                86
            ],
            "revisions": [
                87,
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
                110
            ],
            "segments": [
                123
            ],
            "subtitles": [
                128
            ],
            "scripts": [
                131
            ],
            "clips": [
                125
            ],
            "comments": [
                133
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
                102
            ],
            "matchTerms": [
                1
            ],
            "cover": [
                92
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
        "CategoryTile": {
            "category": [
                39
            ],
            "count": [
                3
            ],
            "preview": [
                92
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
            "matchTerms": [
                1
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
                42
            ],
            "details": [
                1
            ],
            "cover": [
                92
            ],
            "itemCount": [
                3
            ],
            "items": [
                44,
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
                148
            ],
            "subjectId": [
                2
            ],
            "record": [
                153
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
                148
            ],
            "fieldType": [
                45
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
                148
            ],
            "fieldType": [
                45
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
                59,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "files": [
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
                92
            ],
            "status": [
                150
            ],
            "studio": [
                101
            ],
            "performers": [
                96
            ],
            "tags": [
                102
            ],
            "imageCount": [
                3
            ],
            "images": [
                95
            ],
            "video": [
                38
            ],
            "sources": [
                79
            ],
            "userState": [
                154
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                86
            ],
            "revisions": [
                87,
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
                148
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
                151
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
                51
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
                147
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
                49
            ],
            "people": [
                54
            ],
            "guesses": [
                58
            ],
            "__typename": [
                1
            ]
        },
        "FacePerson": {
            "sightings": [
                55
            ],
            "candidates": [
                56
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
                147
            ],
            "__typename": [
                1
            ]
        },
        "FaceCandidate": {
            "performer": [
                96
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
        "FaceMatch": {
            "video": [
                38
            ],
            "similarity": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "FaceGuess": {
            "performer": [
                96
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
                152
            ],
            "record": [
                153
            ],
            "score": [
                5
            ],
            "evidence": [
                60
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
                92
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
                59
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
                63
            ],
            "performers": [
                63
            ],
            "studios": [
                63
            ],
            "sites": [
                63
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
                92
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
                59
            ],
            "failures": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "MergeJournal": {
            "id": [
                2
            ],
            "createdAt": [
                1
            ],
            "subjectKind": [
                148
            ],
            "keptId": [
                2
            ],
            "mergedId": [
                2
            ],
            "mergedName": [
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
                148
            ],
            "first": [
                153
            ],
            "second": [
                153
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
                146
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
                148
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
                146
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
                142
            ],
            "file": [
                71
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
                73
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
                75
            ],
            "model": [
                1
            ],
            "atSeconds": [
                5
            ],
            "region": [
                147
            ],
            "__typename": [
                1
            ]
        },
        "EmbeddingInput": {
            "subject": [
                151
            ],
            "kind": [
                75
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
                147
            ],
            "__typename": [
                1
            ]
        },
        "Audio": {
            "files": [
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
                92
            ],
            "status": [
                150
            ],
            "studio": [
                101
            ],
            "performers": [
                96
            ],
            "tags": [
                102
            ],
            "durationSeconds": [
                5
            ],
            "sources": [
                79
            ],
            "streams": [
                103
            ],
            "userState": [
                154
            ],
            "fieldSources": [
                86
            ],
            "revisions": [
                87,
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
                123
            ],
            "subtitles": [
                128
            ],
            "scripts": [
                131
            ],
            "comments": [
                133
            ],
            "__typename": [
                1
            ]
        },
        "Source": {
            "snapshot": [
                92
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
                92
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
                91
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
                150
            ],
            "vr": [
                104
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
                83
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
                88
            ],
            "source": [
                79
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
                148
            ],
            "subjectId": [
                2
            ],
            "field": [
                1
            ],
            "before": [
                147
            ],
            "after": [
                147
            ],
            "writtenBy": [
                88
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
                102
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
                48
            ],
            "performers": [
                96
            ],
            "studios": [
                101
            ],
            "tags": [
                102
            ],
            "audios": [
                78
            ],
            "stories": [
                107
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
                105
            ],
            "stereo": [
                106
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
                92
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
                92
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
                92
            ],
            "details": [
                1
            ],
            "attributes": [
                100
            ],
            "links": [
                1
            ],
            "tags": [
                102
            ],
            "sources": [
                79
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                154
            ],
            "fieldSources": [
                86
            ],
            "revisions": [
                87,
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
                99
            ],
            "__typename": [
                1
            ]
        },
        "PerformerSuggestion": {
            "performer": [
                96
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
                96
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
                92
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                101
            ],
            "tags": [
                102
            ],
            "sources": [
                79
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                154
            ],
            "fieldSources": [
                86
            ],
            "revisions": [
                87,
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
                154
            ],
            "revisions": [
                87,
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
                79
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
                105
            ],
            "stereo": [
                106
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
                92
            ],
            "status": [
                150
            ],
            "studio": [
                101
            ],
            "performers": [
                96
            ],
            "tags": [
                102
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
                108
            ],
            "sources": [
                79
            ],
            "userState": [
                154
            ],
            "fieldSources": [
                86
            ],
            "revisions": [
                87,
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
            "parent": [
                2
            ],
            "choice": [
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
                92
            ],
            "status": [
                150
            ],
            "creator": [
                96
            ],
            "text": [
                1
            ],
            "tags": [
                102
            ],
            "attachments": [
                153
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
                79
            ],
            "userState": [
                154
            ],
            "fieldSources": [
                86
            ],
            "revisions": [
                87,
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
                92
            ],
            "studio": [
                101
            ],
            "videos": [
                38
            ],
            "sources": [
                79
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
                111
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
                113
            ],
            "expiresAt": [
                1
            ],
            "lastCheckedAt": [
                1
            ],
            "storesLogin": [
                4
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
                117
            ],
            "items": [
                118
            ],
            "actions": [
                119
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
                153
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
                116
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
                115
            ],
            "account": [
                114
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
                136
            ],
            "settingValues": [
                137
            ],
            "login": [
                135
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
                148
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
                102
            ],
            "performers": [
                96
            ],
            "origin": [
                122
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
                92
            ],
            "file": [
                71
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
                92
            ],
            "origin": [
                127
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
                127
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
                130
            ],
            "asset": [
                92
            ],
            "axis": [
                1
            ],
            "writtenBy": [
                88
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
                79
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
                96
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
        "PluginLogin": {
            "fields": [
                136
            ],
            "cookies": [
                4
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
                148
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
                140
            ],
            "status": [
                141
            ],
            "subject": [
                152
            ],
            "plugin": [
                1
            ],
            "payload": [
                147
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
                143
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
                144
            ],
            "kind": [
                145
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
                148
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
                148
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
                48
            ],
            "on_Audio": [
                78
            ],
            "on_Story": [
                107
            ],
            "on_Post": [
                109
            ],
            "on_Performer": [
                96
            ],
            "on_Studio": [
                101
            ],
            "on_Tag": [
                102
            ],
            "on_Group": [
                110
            ],
            "on_Category": [
                39
            ],
            "on_Collection": [
                43
            ],
            "on_CamModel": [
                35
            ],
            "on_Segment": [
                123
            ],
            "on_Clip": [
                125
            ],
            "__typename": [
                1
            ]
        },
        "UserState": {
            "subjectKind": [
                148
            ],
            "subjectId": [
                2
            ],
            "record": [
                153
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
                147
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
                148
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
                156
            ],
            "subjectKind": [
                148
            ],
            "subjectId": [
                2
            ],
            "data": [
                147
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
                156
            ],
            "subject": [
                151
            ],
            "data": [
                147
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
                156
            ],
            "subject": [
                151
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
                161
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
                161
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
                147
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
                147
            ],
            "__typename": [
                1
            ]
        },
        "Filter": {
            "kinds": [
                149
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
                150
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
                164
            ],
            "sort": [
                162
            ],
            "__typename": [
                1
            ]
        },
        "FilterInput": {
            "kinds": [
                149
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
                150
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
                165
            ],
            "sort": [
                163
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
                166
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