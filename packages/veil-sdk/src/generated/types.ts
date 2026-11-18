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
        34,
        44,
        47,
        75,
        77,
        90,
        91,
        108,
        109,
        116,
        125,
        130,
        133,
        144,
        145,
        147,
        149,
        151,
        152,
        153,
        154,
        160,
        165
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
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "newSince": [
                157,
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
                35,
                {
                    "query": [
                        33
                    ]
                }
            ],
            "camRecommendations": [
                35,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "camModels": [
                36,
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
                36,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "activeRecordings": [
                39
            ],
            "categories": [
                41
            ],
            "category": [
                41,
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
                42,
                {
                    "kind": [
                        153,
                        "MediaKind!"
                    ]
                }
            ],
            "collections": [
                45
            ],
            "collection": [
                45,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "builtInCollection": [
                45,
                {
                    "kind": [
                        44,
                        "CollectionKind!"
                    ]
                }
            ],
            "collectionsWith": [
                45,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "fieldDefs": [
                48,
                {
                    "kind": [
                        152
                    ]
                }
            ],
            "faceScans": [
                51,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "facePictures": [
                54,
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
                55,
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
                59,
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
                61,
                {
                    "filter": [
                        171
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
                61,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "shorts": [
                63,
                {
                    "filter": [
                        171
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "taste": [
                64
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
                69,
                {
                    "kind": [
                        152,
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
                68,
                {
                    "kind": [
                        152,
                        "RecordKind!"
                    ],
                    "keptId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "libraryRoots": [
                70
            ],
            "libraryRoot": [
                70,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "planOrganize": [
                71,
                {
                    "rootId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "downloads": [
                74
            ],
            "sameMedia": [
                157,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "nearest": [
                61,
                {
                    "kind": [
                        77,
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
                        152,
                        "[RecordKind!]"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "similar": [
                61,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "kind": [
                        77,
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
                40,
                {
                    "filter": [
                        171
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
                        171
                    ]
                }
            ],
            "videoTagCounts": [
                92,
                {
                    "filter": [
                        171
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "video": [
                40,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "galleries": [
                50,
                {
                    "filter": [
                        171
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
                        171
                    ]
                }
            ],
            "gallery": [
                50,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "performers": [
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
            "performer": [
                98,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "performerDuplicates": [
                98,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "collecting": [
                100
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
                103,
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
                103,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "tags": [
                104,
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
                104,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "media": [
                157,
                {
                    "filter": [
                        171
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
                        171
                    ]
                }
            ],
            "mediaTagCounts": [
                92,
                {
                    "kind": [
                        153,
                        "MediaKind!"
                    ],
                    "filter": [
                        171
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "creditedPerformers": [
                98,
                {
                    "kind": [
                        153,
                        "MediaKind!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "audios": [
                80,
                {
                    "filter": [
                        171
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
                        171
                    ]
                }
            ],
            "audio": [
                80,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "similarAudios": [
                80,
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
                110,
                {
                    "filter": [
                        171
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
                        171
                    ]
                }
            ],
            "story": [
                110,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "similarStories": [
                110,
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
                112,
                {
                    "filter": [
                        171
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
                        171
                    ]
                }
            ],
            "post": [
                112,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "groups": [
                113,
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
                113,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "settings": [
                151
            ],
            "apiTokens": [
                114
            ],
            "accounts": [
                117
            ],
            "contributions": [
                118,
                {
                    "kind": [
                        1
                    ]
                }
            ],
            "pluginView": [
                119,
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
                        151
                    ]
                }
            ],
            "clips": [
                128,
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
                128,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                124
            ],
            "plugin": [
                124,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "pluginPackages": [
                137,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "trash": [
                143,
                {
                    "kind": [
                        152,
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
                146,
                {
                    "statuses": [
                        145,
                        "[JobStatus!]"
                    ],
                    "kinds": [
                        144,
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
                146,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "devices": [
                148
            ],
            "currentDevice": [
                148
            ],
            "storageLocations": [
                150,
                {
                    "deviceId": [
                        2
                    ]
                }
            ],
            "userStates": [
                158,
                {
                    "filter": [
                        159
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
                158,
                {
                    "kind": [
                        152
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "liked": [
                157,
                {
                    "kind": [
                        152,
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
            "events": [
                162,
                {
                    "filter": [
                        164
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
                        164
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
                161
            ],
            "savedFilters": [
                172
            ],
            "savedFilter": [
                172,
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
                        155,
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
                146,
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
                        151
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
                        151
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
                        155,
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
                    ],
                    "preview": [
                        4
                    ]
                }
            ],
            "setCamAutoRecord": [
                36,
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
                36,
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
                146,
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
                41,
                {
                    "input": [
                        43,
                        "CategoryInput!"
                    ]
                }
            ],
            "editCategory": [
                41,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        43,
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
                41,
                {
                    "ids": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "createCollection": [
                45,
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
                45,
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
                45,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subjects": [
                        155,
                        "[RecordRefInput!]!"
                    ],
                    "position": [
                        3
                    ]
                }
            ],
            "removeFromCollection": [
                45,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subjects": [
                        155,
                        "[RecordRefInput!]!"
                    ]
                }
            ],
            "moveInCollection": [
                45,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "position": [
                        3,
                        "Int!"
                    ]
                }
            ],
            "clearCollection": [
                45,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "saveFieldDef": [
                48,
                {
                    "id": [
                        2
                    ],
                    "input": [
                        49,
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
                158,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "fieldDefId": [
                        2,
                        "ID!"
                    ],
                    "value": [
                        151
                    ]
                }
            ],
            "recordFaceScan": [
                51,
                {
                    "input": [
                        52,
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
                67,
                {
                    "input": [
                        66,
                        "LookupInput!"
                    ]
                }
            ],
            "merge": [
                157,
                {
                    "kind": [
                        152,
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
                157,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "mergeAlike": [
                157,
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
                70,
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
                70,
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
                146,
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
                        72,
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
                146,
                {
                    "subject": [
                        155,
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
                76,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "fileId": [
                        2
                    ],
                    "algorithm": [
                        75,
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
                146,
                {
                    "fileId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "snapshotSource": [
                146,
                {
                    "sourceId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addEmbedding": [
                78,
                {
                    "input": [
                        79,
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
                40,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "fetchGallery": [
                50,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "fetchPerformer": [
                98,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "refresh": [
                        4
                    ],
                    "plugins": [
                        1,
                        "[String!]"
                    ]
                }
            ],
            "unassignPerformer": [
                98,
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
                98,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "createStudio": [
                103,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "browse": [
                93,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "kind": [
                        91,
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
            "fetchUploader": [
                40,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "url": [
                        1,
                        "String!"
                    ],
                    "name": [
                        1,
                        "String!"
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
                93,
                {
                    "kind": [
                        91,
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
                40,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        83,
                        "MediaEdit!"
                    ]
                }
            ],
            "editGallery": [
                50,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        83,
                        "MediaEdit!"
                    ]
                }
            ],
            "editPerformer": [
                98,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        84,
                        "PerformerEdit!"
                    ]
                }
            ],
            "editStudio": [
                103,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        86,
                        "StudioEdit!"
                    ]
                }
            ],
            "editTag": [
                104,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        87,
                        "TagEdit!"
                    ]
                }
            ],
            "lockField": [
                88,
                {
                    "subject": [
                        155,
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
                157,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "edit": [
                        83,
                        "MediaEdit!"
                    ]
                }
            ],
            "fetchMedia": [
                157,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "readChapter": [
                111,
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
                151,
                {
                    "values": [
                        151,
                        "JSON!"
                    ]
                }
            ],
            "createApiToken": [
                115,
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
                117,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "values": [
                        151,
                        "JSON!"
                    ]
                }
            ],
            "importAccountCookies": [
                117,
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
                117,
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
                123,
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
                        151
                    ],
                    "input": [
                        151
                    ]
                }
            ],
            "addSegment": [
                126,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "input": [
                        127,
                        "SegmentInput!"
                    ]
                }
            ],
            "editSegment": [
                126,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        127,
                        "SegmentInput!"
                    ]
                }
            ],
            "addClip": [
                128,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        129,
                        "ClipInput!"
                    ]
                }
            ],
            "editClip": [
                128,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        129,
                        "ClipInput!"
                    ]
                }
            ],
            "attachClipFile": [
                128,
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
                131,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "input": [
                        132,
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
                134,
                {
                    "subject": [
                        155,
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
                40,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "trickplay": [
                        135
                    ]
                }
            ],
            "reportStreamSpeed": [
                106,
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
                157,
                {
                    "subject": [
                        155,
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
                        142,
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
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "restoreRecord": [
                4,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "purgeRecord": [
                4,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "emptyTrash": [
                3
            ],
            "queueFetch": [
                146,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ]
                }
            ],
            "cancelJob": [
                146,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "retryJob": [
                146,
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
                148,
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
                150,
                {
                    "kind": [
                        149,
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
                150,
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
                158,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "rating": [
                        5
                    ]
                }
            ],
            "setFavorite": [
                158,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "favorite": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setBlocked": [
                158,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "blocked": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setNotes": [
                158,
                {
                    "subject": [
                        155,
                        "RecordRefInput!"
                    ],
                    "notes": [
                        1,
                        "String!"
                    ]
                }
            ],
            "recordEvent": [
                162,
                {
                    "event": [
                        163,
                        "EventInput!"
                    ]
                }
            ],
            "recordEvents": [
                3,
                {
                    "events": [
                        163,
                        "[EventInput!]!"
                    ]
                }
            ],
            "clearEvents": [
                3,
                {
                    "kinds": [
                        160,
                        "[EventKind!]"
                    ]
                }
            ],
            "saveFilter": [
                172,
                {
                    "id": [
                        2
                    ],
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        171,
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
                82,
                {
                    "kind": [
                        91,
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
                        171
                    ],
                    "content": [
                        1
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
                170
            ],
            "targetKind": [
                152
            ],
            "targetId": [
                2
            ],
            "target": [
                157
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
                171
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
                171
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
                156
            ],
            "record": [
                157
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
                151
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
                170
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
                171
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
                151
            ],
            "condition": [
                170
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
                151
            ],
            "condition": [
                171
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
                151
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
                151
            ],
            "error": [
                1
            ],
            "touched": [
                156
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
            "sort": [
                34
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
        "CamSort": {},
        "CamRoomPage": {
            "rooms": [
                36
            ],
            "total": [
                3
            ],
            "reasons": [
                1
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
                98
            ],
            "thumbnail": [
                94
            ],
            "tags": [
                1
            ],
            "gender": [
                1
            ],
            "live": [
                38
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
                40
            ],
            "shows": [
                37
            ],
            "userState": [
                158
            ],
            "__typename": [
                1
            ]
        },
        "CamShow": {
            "startedAt": [
                1
            ],
            "endedAt": [
                1
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
                94
            ],
            "__typename": [
                1
            ]
        },
        "ActiveRecording": {
            "job": [
                146
            ],
            "camModel": [
                36
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
                36
            ],
            "camRooms": [
                36
            ],
            "categories": [
                41
            ],
            "related": [
                61,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "siteRelated": [
                40,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "files": [
                73
            ],
            "fingerprints": [
                76
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
                94
            ],
            "status": [
                154
            ],
            "studio": [
                103
            ],
            "performers": [
                98
            ],
            "suggestedPerformers": [
                99
            ],
            "tags": [
                104
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
                95
            ],
            "preview": [
                94
            ],
            "trickplay": [
                96
            ],
            "previewFrames": [
                94
            ],
            "gallery": [
                50
            ],
            "sources": [
                81
            ],
            "streams": [
                106
            ],
            "userState": [
                158
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                88
            ],
            "revisions": [
                89,
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
                113
            ],
            "segments": [
                126
            ],
            "subtitles": [
                131
            ],
            "scripts": [
                134
            ],
            "clips": [
                128
            ],
            "comments": [
                136
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
                41
            ],
            "children": [
                41
            ],
            "matchTags": [
                104
            ],
            "matchTerms": [
                1
            ],
            "cover": [
                94
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
                41
            ],
            "count": [
                3
            ],
            "preview": [
                94
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
                44
            ],
            "details": [
                1
            ],
            "cover": [
                94
            ],
            "itemCount": [
                3
            ],
            "items": [
                46,
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
                152
            ],
            "subjectId": [
                2
            ],
            "record": [
                157
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
                152
            ],
            "fieldType": [
                47
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
                152
            ],
            "fieldType": [
                47
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
                41
            ],
            "related": [
                61,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "files": [
                73
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
                94
            ],
            "status": [
                154
            ],
            "studio": [
                103
            ],
            "performers": [
                98
            ],
            "tags": [
                104
            ],
            "imageCount": [
                3
            ],
            "images": [
                97
            ],
            "video": [
                40
            ],
            "sources": [
                81
            ],
            "userState": [
                158
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                88
            ],
            "revisions": [
                89,
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
                152
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
                155
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
                53
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
                151
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
                51
            ],
            "people": [
                56
            ],
            "guesses": [
                60
            ],
            "__typename": [
                1
            ]
        },
        "FacePerson": {
            "sightings": [
                57
            ],
            "candidates": [
                58
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
                151
            ],
            "__typename": [
                1
            ]
        },
        "FaceCandidate": {
            "performer": [
                98
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
                40
            ],
            "__typename": [
                1
            ]
        },
        "FaceMatch": {
            "video": [
                40
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
                98
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
                156
            ],
            "record": [
                157
            ],
            "score": [
                5
            ],
            "evidence": [
                62
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
                94
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
                61
            ],
            "video": [
                40
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
                65
            ],
            "performers": [
                65
            ],
            "studios": [
                65
            ],
            "sites": [
                65
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
                94
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
                61
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
                152
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
                152
            ],
            "first": [
                157
            ],
            "second": [
                157
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
                150
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
                152
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
                150
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
                146
            ],
            "file": [
                73
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
                75
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
                77
            ],
            "model": [
                1
            ],
            "atSeconds": [
                5
            ],
            "region": [
                151
            ],
            "__typename": [
                1
            ]
        },
        "EmbeddingInput": {
            "subject": [
                155
            ],
            "kind": [
                77
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
                151
            ],
            "__typename": [
                1
            ]
        },
        "Audio": {
            "files": [
                73
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
                94
            ],
            "status": [
                154
            ],
            "studio": [
                103
            ],
            "performers": [
                98
            ],
            "tags": [
                104
            ],
            "durationSeconds": [
                5
            ],
            "sources": [
                81
            ],
            "streams": [
                106
            ],
            "userState": [
                158
            ],
            "fieldSources": [
                88
            ],
            "revisions": [
                89,
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
                126
            ],
            "subtitles": [
                131
            ],
            "scripts": [
                134
            ],
            "comments": [
                136
            ],
            "__typename": [
                1
            ]
        },
        "Source": {
            "snapshot": [
                94
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
                94
            ],
            "fetchedAt": [
                1
            ],
            "uploader": [
                105
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
                93
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
                154
            ],
            "vr": [
                107
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
                85
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
                90
            ],
            "source": [
                81
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
                152
            ],
            "subjectId": [
                2
            ],
            "field": [
                1
            ],
            "before": [
                151
            ],
            "after": [
                151
            ],
            "writtenBy": [
                90
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
                104
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
                40
            ],
            "galleries": [
                50
            ],
            "performers": [
                98
            ],
            "studios": [
                103
            ],
            "tags": [
                104
            ],
            "audios": [
                80
            ],
            "stories": [
                110
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
                108
            ],
            "stereo": [
                109
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
                94
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
                94
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
                94
            ],
            "details": [
                1
            ],
            "attributes": [
                102
            ],
            "links": [
                1
            ],
            "tags": [
                104
            ],
            "sources": [
                81
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                158
            ],
            "fieldSources": [
                88
            ],
            "revisions": [
                89,
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
                101
            ],
            "__typename": [
                1
            ]
        },
        "PerformerSuggestion": {
            "performer": [
                98
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
                98
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
                94
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                103
            ],
            "tags": [
                104
            ],
            "sources": [
                81
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                158
            ],
            "fieldSources": [
                88
            ],
            "revisions": [
                89,
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
                158
            ],
            "revisions": [
                89,
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
        "Uploader": {
            "name": [
                1
            ],
            "url": [
                1
            ],
            "externalId": [
                1
            ],
            "image": [
                94
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
                81
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
                108
            ],
            "stereo": [
                109
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
                94
            ],
            "status": [
                154
            ],
            "studio": [
                103
            ],
            "performers": [
                98
            ],
            "tags": [
                104
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
                111
            ],
            "sources": [
                81
            ],
            "userState": [
                158
            ],
            "fieldSources": [
                88
            ],
            "revisions": [
                89,
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
                94
            ],
            "status": [
                154
            ],
            "creator": [
                98
            ],
            "text": [
                1
            ],
            "tags": [
                104
            ],
            "attachments": [
                157
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
                81
            ],
            "userState": [
                158
            ],
            "fieldSources": [
                88
            ],
            "revisions": [
                89,
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
                94
            ],
            "studio": [
                103
            ],
            "videos": [
                40
            ],
            "sources": [
                81
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
                114
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
                116
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
                120
            ],
            "items": [
                121
            ],
            "actions": [
                122
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
                157
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
                119
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
                118
            ],
            "account": [
                117
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
                140
            ],
            "settingValues": [
                141
            ],
            "login": [
                139
            ],
            "sites": [
                138
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
                152
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
                104
            ],
            "performers": [
                98
            ],
            "origin": [
                125
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
                40
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
                94
            ],
            "file": [
                73
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
                94
            ],
            "origin": [
                130
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
                130
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
                133
            ],
            "asset": [
                94
            ],
            "axis": [
                1
            ],
            "writtenBy": [
                90
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
                81
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
                98
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
        "PluginSite": {
            "name": [
                1
            ],
            "host": [
                1
            ],
            "iconUrl": [
                1
            ],
            "content": [
                1
            ],
            "offByDefault": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "PluginLogin": {
            "fields": [
                140
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
                152
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
                144
            ],
            "status": [
                145
            ],
            "subject": [
                156
            ],
            "plugin": [
                1
            ],
            "payload": [
                151
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
                147
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
                148
            ],
            "kind": [
                149
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
                152
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
                152
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
                40
            ],
            "on_Gallery": [
                50
            ],
            "on_Audio": [
                80
            ],
            "on_Story": [
                110
            ],
            "on_Post": [
                112
            ],
            "on_Performer": [
                98
            ],
            "on_Studio": [
                103
            ],
            "on_Tag": [
                104
            ],
            "on_Group": [
                113
            ],
            "on_Category": [
                41
            ],
            "on_Collection": [
                45
            ],
            "on_CamModel": [
                36
            ],
            "on_Segment": [
                126
            ],
            "on_Clip": [
                128
            ],
            "__typename": [
                1
            ]
        },
        "UserState": {
            "subjectKind": [
                152
            ],
            "subjectId": [
                2
            ],
            "record": [
                157
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
                151
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
                152
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
                160
            ],
            "subjectKind": [
                152
            ],
            "subjectId": [
                2
            ],
            "data": [
                151
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
                160
            ],
            "subject": [
                155
            ],
            "data": [
                151
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
                160
            ],
            "subject": [
                155
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
                165
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
                165
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
                151
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
                151
            ],
            "__typename": [
                1
            ]
        },
        "Filter": {
            "kinds": [
                153
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
                154
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
                168
            ],
            "sort": [
                166
            ],
            "__typename": [
                1
            ]
        },
        "FilterInput": {
            "kinds": [
                153
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
                154
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
                169
            ],
            "sort": [
                167
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
                170
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