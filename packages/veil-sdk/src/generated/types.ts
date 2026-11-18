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
        76,
        78,
        91,
        92,
        109,
        110,
        117,
        126,
        133,
        136,
        147,
        148,
        150,
        152,
        154,
        155,
        156,
        157,
        163,
        168
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
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "newSince": [
                160,
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
                        156,
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
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "fieldDefs": [
                48,
                {
                    "kind": [
                        155
                    ]
                }
            ],
            "faceScans": [
                51,
                {
                    "subject": [
                        158,
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
                        174
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
                        158,
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
                        174
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
                70,
                {
                    "kind": [
                        155,
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
                69,
                {
                    "kind": [
                        155,
                        "RecordKind!"
                    ],
                    "keptId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "libraryRoots": [
                71
            ],
            "libraryRoot": [
                71,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "planOrganize": [
                72,
                {
                    "rootId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "downloads": [
                75
            ],
            "sameMedia": [
                160,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "nearest": [
                61,
                {
                    "kind": [
                        78,
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
                        155,
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
                        158,
                        "RecordRefInput!"
                    ],
                    "kind": [
                        78,
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
                        174
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
                        174
                    ]
                }
            ],
            "videoTagCounts": [
                93,
                {
                    "filter": [
                        174
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
                        174
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
                        174
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
            "performer": [
                99,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "performerDuplicates": [
                99,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "collecting": [
                101
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
            "studio": [
                104,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "tags": [
                105,
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
                105,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "media": [
                160,
                {
                    "filter": [
                        174
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
                        174
                    ]
                }
            ],
            "mediaTagCounts": [
                93,
                {
                    "kind": [
                        156,
                        "MediaKind!"
                    ],
                    "filter": [
                        174
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "creditedPerformers": [
                99,
                {
                    "kind": [
                        156,
                        "MediaKind!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "audios": [
                81,
                {
                    "filter": [
                        174
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
                        174
                    ]
                }
            ],
            "audio": [
                81,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "similarAudios": [
                81,
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
                111,
                {
                    "filter": [
                        174
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
                        174
                    ]
                }
            ],
            "story": [
                111,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "similarStories": [
                111,
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
                113,
                {
                    "filter": [
                        174
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
                        174
                    ]
                }
            ],
            "post": [
                113,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "groups": [
                114,
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
                114,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "settings": [
                154
            ],
            "apiTokens": [
                115
            ],
            "accounts": [
                118
            ],
            "contributions": [
                119,
                {
                    "kind": [
                        1
                    ]
                }
            ],
            "pluginView": [
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
                    "input": [
                        154
                    ]
                }
            ],
            "clips": [
                131,
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
                131,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "sceneMatches": [
                130,
                {
                    "tags": [
                        1,
                        "[String!]!"
                    ],
                    "videoIds": [
                        2,
                        "[ID!]"
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "sceneLabelCounts": [
                93,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "likedSceneTags": [
                105
            ],
            "plugins": [
                125
            ],
            "plugin": [
                125,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "pluginPackages": [
                140,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "trash": [
                146,
                {
                    "kind": [
                        155,
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
                149,
                {
                    "statuses": [
                        148,
                        "[JobStatus!]"
                    ],
                    "kinds": [
                        147,
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
                149,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "devices": [
                151
            ],
            "currentDevice": [
                151
            ],
            "storageLocations": [
                153,
                {
                    "deviceId": [
                        2
                    ]
                }
            ],
            "userStates": [
                161,
                {
                    "filter": [
                        162
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
                161,
                {
                    "kind": [
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
            "liked": [
                160,
                {
                    "kind": [
                        155,
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
                165,
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
            "eventCount": [
                3,
                {
                    "filter": [
                        167
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
                164
            ],
            "savedFilters": [
                175
            ],
            "savedFilter": [
                175,
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
                        158,
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
                149,
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
                        154
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
                        154
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
                        158,
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
                149,
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
                        158,
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
                        158,
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
                        158,
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
                161,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "fieldDefId": [
                        2,
                        "ID!"
                    ],
                    "value": [
                        154
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
                68,
                {
                    "input": [
                        67,
                        "LookupInput!"
                    ]
                }
            ],
            "merge": [
                160,
                {
                    "kind": [
                        155,
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
                160,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "mergeAlike": [
                160,
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
                71,
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
                71,
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
                149,
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
                        73,
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
                149,
                {
                    "subject": [
                        158,
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
                77,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "fileId": [
                        2
                    ],
                    "algorithm": [
                        76,
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
                149,
                {
                    "fileId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "snapshotSource": [
                149,
                {
                    "sourceId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addEmbedding": [
                79,
                {
                    "input": [
                        80,
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
                99,
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
                99,
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
                99,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "createStudio": [
                104,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "browse": [
                94,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "kind": [
                        92,
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
                94,
                {
                    "kind": [
                        92,
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
                        84,
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
                        84,
                        "MediaEdit!"
                    ]
                }
            ],
            "editPerformer": [
                99,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        85,
                        "PerformerEdit!"
                    ]
                }
            ],
            "editStudio": [
                104,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        87,
                        "StudioEdit!"
                    ]
                }
            ],
            "editTag": [
                105,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "edit": [
                        88,
                        "TagEdit!"
                    ]
                }
            ],
            "lockField": [
                89,
                {
                    "subject": [
                        158,
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
                160,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "edit": [
                        84,
                        "MediaEdit!"
                    ]
                }
            ],
            "fetchMedia": [
                160,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "readChapter": [
                112,
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
                154,
                {
                    "values": [
                        154,
                        "JSON!"
                    ]
                }
            ],
            "createApiToken": [
                116,
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
                118,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "values": [
                        154,
                        "JSON!"
                    ]
                }
            ],
            "importAccountCookies": [
                118,
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
                118,
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
                124,
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
                        154
                    ],
                    "input": [
                        154
                    ]
                }
            ],
            "addSegment": [
                127,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "input": [
                        128,
                        "SegmentInput!"
                    ]
                }
            ],
            "editSegment": [
                127,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        128,
                        "SegmentInput!"
                    ]
                }
            ],
            "replaceAiSegments": [
                127,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "segments": [
                        129,
                        "[AiSegmentInput!]!"
                    ]
                }
            ],
            "addClip": [
                131,
                {
                    "videoId": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        132,
                        "ClipInput!"
                    ]
                }
            ],
            "editClip": [
                131,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "input": [
                        132,
                        "ClipInput!"
                    ]
                }
            ],
            "attachClipFile": [
                131,
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
                134,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "input": [
                        135,
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
                137,
                {
                    "subject": [
                        158,
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
                        138
                    ]
                }
            ],
            "reportStreamSpeed": [
                107,
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
                160,
                {
                    "subject": [
                        158,
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
                        145,
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
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "restoreRecord": [
                4,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "purgeRecord": [
                4,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "emptyTrash": [
                3
            ],
            "queueFetch": [
                149,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ]
                }
            ],
            "cancelJob": [
                149,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "retryJob": [
                149,
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
                151,
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
                153,
                {
                    "kind": [
                        152,
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
                153,
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
                161,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "rating": [
                        5
                    ]
                }
            ],
            "setFavorite": [
                161,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "favorite": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setBlocked": [
                161,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "blocked": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "setNotes": [
                161,
                {
                    "subject": [
                        158,
                        "RecordRefInput!"
                    ],
                    "notes": [
                        1,
                        "String!"
                    ]
                }
            ],
            "recordEvent": [
                165,
                {
                    "event": [
                        166,
                        "EventInput!"
                    ]
                }
            ],
            "recordEvents": [
                3,
                {
                    "events": [
                        166,
                        "[EventInput!]!"
                    ]
                }
            ],
            "clearEvents": [
                3,
                {
                    "kinds": [
                        163,
                        "[EventKind!]"
                    ]
                }
            ],
            "saveFilter": [
                175,
                {
                    "id": [
                        2
                    ],
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        174,
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
                83,
                {
                    "kind": [
                        92,
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
                        174
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
                173
            ],
            "targetKind": [
                155
            ],
            "targetId": [
                2
            ],
            "target": [
                160
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
                174
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
                174
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
                159
            ],
            "record": [
                160
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
                154
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
                173
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
                174
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
                154
            ],
            "condition": [
                173
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
                154
            ],
            "condition": [
                174
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
                154
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
                154
            ],
            "error": [
                1
            ],
            "touched": [
                159
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
                99
            ],
            "thumbnail": [
                95
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
                161
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
                95
            ],
            "__typename": [
                1
            ]
        },
        "ActiveRecording": {
            "job": [
                149
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
                74
            ],
            "fingerprints": [
                77
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
                95
            ],
            "status": [
                157
            ],
            "studio": [
                104
            ],
            "performers": [
                99
            ],
            "suggestedPerformers": [
                100
            ],
            "tags": [
                105
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
                96
            ],
            "preview": [
                95
            ],
            "trickplay": [
                97
            ],
            "previewFrames": [
                95
            ],
            "gallery": [
                50
            ],
            "sources": [
                82
            ],
            "streams": [
                107
            ],
            "copies": [
                40
            ],
            "userState": [
                161
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                89
            ],
            "revisions": [
                90,
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
                114
            ],
            "segments": [
                127
            ],
            "subtitles": [
                134
            ],
            "scripts": [
                137
            ],
            "clips": [
                131
            ],
            "comments": [
                139
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
                105
            ],
            "matchTerms": [
                1
            ],
            "cover": [
                95
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
                95
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
                95
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
                155
            ],
            "subjectId": [
                2
            ],
            "record": [
                160
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
                155
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
                155
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
                95
            ],
            "status": [
                157
            ],
            "studio": [
                104
            ],
            "performers": [
                99
            ],
            "tags": [
                105
            ],
            "imageCount": [
                3
            ],
            "images": [
                98
            ],
            "video": [
                40
            ],
            "sources": [
                82
            ],
            "userState": [
                161
            ],
            "oCount": [
                3
            ],
            "fieldSources": [
                89
            ],
            "revisions": [
                90,
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
                155
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
                158
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
                154
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
                154
            ],
            "__typename": [
                1
            ]
        },
        "FaceCandidate": {
            "performer": [
                99
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
                99
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
                159
            ],
            "record": [
                160
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
                95
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
            "sceneLabels": [
                66
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
                95
            ],
            "weight": [
                5
            ],
            "confirmed": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "SceneLabelTaste": {
            "tag": [
                105
            ],
            "score": [
                5
            ],
            "certainty": [
                5
            ],
            "evidence": [
                1
            ],
            "confirmed": [
                4
            ],
            "videoCount": [
                3
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
                155
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
                155
            ],
            "first": [
                160
            ],
            "second": [
                160
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
                153
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
                155
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
                153
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
                149
            ],
            "file": [
                74
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
                76
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
                78
            ],
            "model": [
                1
            ],
            "atSeconds": [
                5
            ],
            "region": [
                154
            ],
            "__typename": [
                1
            ]
        },
        "EmbeddingInput": {
            "subject": [
                158
            ],
            "kind": [
                78
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
                154
            ],
            "__typename": [
                1
            ]
        },
        "Audio": {
            "files": [
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
                95
            ],
            "status": [
                157
            ],
            "studio": [
                104
            ],
            "performers": [
                99
            ],
            "tags": [
                105
            ],
            "durationSeconds": [
                5
            ],
            "sources": [
                82
            ],
            "streams": [
                107
            ],
            "userState": [
                161
            ],
            "fieldSources": [
                89
            ],
            "revisions": [
                90,
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
                127
            ],
            "subtitles": [
                134
            ],
            "scripts": [
                137
            ],
            "comments": [
                139
            ],
            "__typename": [
                1
            ]
        },
        "Source": {
            "snapshot": [
                95
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
                95
            ],
            "fetchedAt": [
                1
            ],
            "uploader": [
                106
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
                94
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
                157
            ],
            "vr": [
                108
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
                86
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
                91
            ],
            "source": [
                82
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
                155
            ],
            "subjectId": [
                2
            ],
            "field": [
                1
            ],
            "before": [
                154
            ],
            "after": [
                154
            ],
            "writtenBy": [
                91
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
                105
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
                99
            ],
            "studios": [
                104
            ],
            "tags": [
                105
            ],
            "audios": [
                81
            ],
            "stories": [
                111
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
                109
            ],
            "stereo": [
                110
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
                95
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
                95
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
                95
            ],
            "details": [
                1
            ],
            "attributes": [
                103
            ],
            "links": [
                1
            ],
            "tags": [
                105
            ],
            "sources": [
                82
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                161
            ],
            "fieldSources": [
                89
            ],
            "revisions": [
                90,
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
                102
            ],
            "__typename": [
                1
            ]
        },
        "PerformerSuggestion": {
            "performer": [
                99
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
                99
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
                95
            ],
            "details": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                104
            ],
            "tags": [
                105
            ],
            "sources": [
                82
            ],
            "videoCount": [
                3
            ],
            "galleryCount": [
                3
            ],
            "userState": [
                161
            ],
            "fieldSources": [
                89
            ],
            "revisions": [
                90,
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
                161
            ],
            "revisions": [
                90,
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
                95
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
                82
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
                109
            ],
            "stereo": [
                110
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
                95
            ],
            "status": [
                157
            ],
            "studio": [
                104
            ],
            "performers": [
                99
            ],
            "tags": [
                105
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
                112
            ],
            "sources": [
                82
            ],
            "userState": [
                161
            ],
            "fieldSources": [
                89
            ],
            "revisions": [
                90,
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
                95
            ],
            "status": [
                157
            ],
            "creator": [
                99
            ],
            "text": [
                1
            ],
            "tags": [
                105
            ],
            "attachments": [
                160
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
                82
            ],
            "userState": [
                161
            ],
            "fieldSources": [
                89
            ],
            "revisions": [
                90,
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
                95
            ],
            "studio": [
                104
            ],
            "videos": [
                40
            ],
            "sources": [
                82
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
                115
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
                117
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
                121
            ],
            "items": [
                122
            ],
            "actions": [
                123
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
                160
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
                120
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
                119
            ],
            "account": [
                118
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
                143
            ],
            "settingValues": [
                144
            ],
            "login": [
                142
            ],
            "sites": [
                141
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
                155
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
                105
            ],
            "performers": [
                99
            ],
            "origin": [
                126
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
        "AiSegmentInput": {
            "startSeconds": [
                5
            ],
            "endSeconds": [
                5
            ],
            "tag": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SceneMatch": {
            "video": [
                40
            ],
            "segments": [
                127
            ],
            "matchingSeconds": [
                5
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
                95
            ],
            "file": [
                74
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
                95
            ],
            "origin": [
                133
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
                133
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
                136
            ],
            "asset": [
                95
            ],
            "axis": [
                1
            ],
            "writtenBy": [
                91
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
                82
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
                99
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
                143
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
                155
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
                147
            ],
            "status": [
                148
            ],
            "subject": [
                159
            ],
            "plugin": [
                1
            ],
            "payload": [
                154
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
                150
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
                151
            ],
            "kind": [
                152
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
                155
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
                155
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
                81
            ],
            "on_Story": [
                111
            ],
            "on_Post": [
                113
            ],
            "on_Performer": [
                99
            ],
            "on_Studio": [
                104
            ],
            "on_Tag": [
                105
            ],
            "on_Group": [
                114
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
                127
            ],
            "on_Clip": [
                131
            ],
            "__typename": [
                1
            ]
        },
        "UserState": {
            "subjectKind": [
                155
            ],
            "subjectId": [
                2
            ],
            "record": [
                160
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
                154
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
                155
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
                163
            ],
            "subjectKind": [
                155
            ],
            "subjectId": [
                2
            ],
            "data": [
                154
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
                163
            ],
            "subject": [
                158
            ],
            "data": [
                154
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
                163
            ],
            "subject": [
                158
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
                168
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
                168
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
                154
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
                154
            ],
            "__typename": [
                1
            ]
        },
        "Filter": {
            "kinds": [
                156
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
                157
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
                171
            ],
            "sort": [
                169
            ],
            "sceneTagIds": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "FilterInput": {
            "kinds": [
                156
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
                157
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
                172
            ],
            "sort": [
                170
            ],
            "sceneTagIds": [
                2
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
                173
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