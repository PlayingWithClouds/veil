export default {
    "scalars": [
        1,
        2,
        3,
        4,
        5,
        32,
        59,
        73
    ],
    "types": {
        "Query": {
            "vpnStatus": [
                8
            ],
            "stream": [
                9,
                {
                    "url": [
                        1,
                        "String!"
                    ],
                    "pluginName": [
                        1
                    ]
                }
            ],
            "checkProviderSpeeds": [
                11,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "blocklist": [
                14
            ],
            "camRooms": [
                16,
                {
                    "sources": [
                        1,
                        "[String!]"
                    ],
                    "tag": [
                        1
                    ],
                    "query": [
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
                    ]
                }
            ],
            "camSources": [
                17
            ],
            "camModel": [
                15,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "followedCamModels": [
                15
            ],
            "camStream": [
                18,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "camRecordings": [
                19,
                {
                    "camModelId": [
                        2
                    ],
                    "activeOnly": [
                        4
                    ]
                }
            ],
            "changes": [
                20,
                {
                    "targetId": [
                        2,
                        "ID!"
                    ],
                    "key": [
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
            "collections": [
                21,
                {
                    "search": [
                        1
                    ],
                    "tagId": [
                        2
                    ],
                    "origin": [
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
            "collection": [
                21,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "collectionMembers": [
                22,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "collectionIdsForMedia": [
                2,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "galleries": [
                28,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ],
                    "search": [
                        1
                    ],
                    "studioId": [
                        2
                    ],
                    "performerId": [
                        2
                    ],
                    "tagId": [
                        2
                    ],
                    "sources": [
                        1,
                        "[String!]"
                    ]
                }
            ],
            "gallery": [
                28,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "findGalleryVideos": [
                58,
                {
                    "galleryId": [
                        2,
                        "ID!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "images": [
                29,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ],
                    "search": [
                        1
                    ],
                    "studioId": [
                        2
                    ],
                    "performerId": [
                        2
                    ],
                    "tagId": [
                        2
                    ],
                    "galleryId": [
                        2
                    ],
                    "sort": [
                        1
                    ]
                }
            ],
            "image": [
                29,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "jobs": [
                30
            ],
            "job": [
                30,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "downloadedScenes": [
                51
            ],
            "mediaCards": [
                31,
                {
                    "ids": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "oCount": [
                3,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "observation": [
                33,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "observations": [
                33,
                {
                    "targetId": [
                        2,
                        "ID!"
                    ],
                    "status": [
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
            "performers": [
                34,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ],
                    "search": [
                        1
                    ],
                    "tagId": [
                        2
                    ],
                    "sort": [
                        1
                    ]
                }
            ],
            "performer": [
                34,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                36
            ],
            "plugin": [
                36,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "pluginPackages": [
                35,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "pluginCategories": [
                40,
                {
                    "plugin": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "recommendations": [
                44,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ],
                    "refresh": [
                        4
                    ],
                    "sources": [
                        1,
                        "[String!]"
                    ],
                    "minDuration": [
                        3
                    ],
                    "maxDuration": [
                        3
                    ],
                    "excludeTagIds": [
                        2,
                        "[ID!]"
                    ]
                }
            ],
            "tasteProfile": [
                42,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "recommendedRows": [
                46,
                {
                    "rowLimit": [
                        3
                    ],
                    "perRow": [
                        3
                    ]
                }
            ],
            "recommendedCategories": [
                47,
                {
                    "categoryLimit": [
                        3
                    ],
                    "perCategory": [
                        3
                    ]
                }
            ],
            "recommendationPair": [
                51
            ],
            "savedFilters": [
                50
            ],
            "scenes": [
                51,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ],
                    "search": [
                        1
                    ],
                    "studioId": [
                        2
                    ],
                    "performerId": [
                        2
                    ],
                    "tagId": [
                        2
                    ],
                    "sort": [
                        1
                    ],
                    "minRating": [
                        5
                    ],
                    "minDuration": [
                        3
                    ],
                    "maxDuration": [
                        3
                    ],
                    "dateFrom": [
                        1
                    ],
                    "dateTo": [
                        1
                    ],
                    "sources": [
                        1,
                        "[String!]"
                    ],
                    "includeTagIds": [
                        2,
                        "[ID!]"
                    ],
                    "excludeTagIds": [
                        2,
                        "[ID!]"
                    ]
                }
            ],
            "scene": [
                51,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "randomScenes": [
                51,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "recommendedFeed": [
                51,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "sceneHeat": [
                53,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "sceneMarkers": [
                56,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "pluginSearch": [
                57,
                {
                    "query": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        3
                    ],
                    "pluginNames": [
                        1,
                        "[String!]"
                    ]
                }
            ],
            "pluginBrowse": [
                57,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "recommendedBrowse": [
                57,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "findAlikeSources": [
                58,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "searchSuggestions": [
                60,
                {
                    "query": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "sourceSuggestions": [
                61,
                {
                    "query": [
                        1,
                        "String!"
                    ],
                    "limit": [
                        3
                    ]
                }
            ],
            "settings": [
                62
            ],
            "settingEntries": [
                66,
                {
                    "scope": [
                        1
                    ]
                }
            ],
            "mediaStreams": [
                68,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "studios": [
                70,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ],
                    "search": [
                        1
                    ],
                    "tagId": [
                        2
                    ],
                    "sort": [
                        1
                    ]
                }
            ],
            "studio": [
                70,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "searchSubscriptions": [
                75
            ],
            "searchSubscription": [
                75,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "subscriptionForTarget": [
                75,
                {
                    "targetId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "subscriptionFeed": [
                72,
                {
                    "filter": [
                        71
                    ],
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "tags": [
                76,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ],
                    "search": [
                        1
                    ],
                    "category": [
                        1
                    ]
                }
            ],
            "tag": [
                76,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "userRating": [
                77,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "userRatings": [
                77,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "watchHistory": [
                79,
                {
                    "limit": [
                        3
                    ],
                    "offset": [
                        3
                    ]
                }
            ],
            "watchHistoryEntry": [
                79,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "watchlist": [
                81
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
                12,
                {
                    "pluginName": [
                        1,
                        "String!"
                    ],
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "queueDownload": [
                2,
                {
                    "url": [
                        1,
                        "String!"
                    ],
                    "title": [
                        1,
                        "String!"
                    ],
                    "pluginName": [
                        1
                    ]
                }
            ],
            "addBlock": [
                14,
                {
                    "kind": [
                        1,
                        "String!"
                    ],
                    "targetId": [
                        2,
                        "ID!"
                    ],
                    "label": [
                        1
                    ]
                }
            ],
            "removeBlock": [
                4,
                {
                    "targetId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "followCamModel": [
                15,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "unfollowCamModel": [
                15,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "startCamRecording": [
                19,
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
            "stopCamRecording": [
                19,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "setCamAutoRecord": [
                15,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "enabled": [
                        4,
                        "Boolean!"
                    ],
                    "quality": [
                        1
                    ]
                }
            ],
            "createCollection": [
                21,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "renameCollection": [
                21,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "deleteCollection": [
                4,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addToCollection": [
                4,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "removeFromCollection": [
                4,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "reorderCollection": [
                4,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "mediaIds": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "setCollectionTags": [
                21,
                {
                    "collectionId": [
                        2,
                        "ID!"
                    ],
                    "tagIds": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "identifyFace": [
                23,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ],
                    "imageId": [
                        2
                    ],
                    "frameBase64": [
                        1
                    ]
                }
            ],
            "identifyPicture": [
                23,
                {
                    "pictureBase64": [
                        1
                    ],
                    "pictureUrl": [
                        1
                    ]
                }
            ],
            "creditPerformerByName": [
                34,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ],
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "ensureGalleryImages": [
                28,
                {
                    "galleryId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "linkGalleryVideo": [
                51,
                {
                    "galleryId": [
                        2,
                        "ID!"
                    ],
                    "pluginName": [
                        1,
                        "String!"
                    ],
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "setImageTags": [
                29,
                {
                    "imageId": [
                        2,
                        "ID!"
                    ],
                    "tagIds": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "deleteJob": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "deleteJobsByKind": [
                3,
                {
                    "kind": [
                        1,
                        "String!"
                    ]
                }
            ],
            "clearJobs": [
                3
            ],
            "reorderDownloads": [
                4,
                {
                    "jobIds": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "retryJob": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "mergeRecords": [
                2,
                {
                    "survivorId": [
                        2,
                        "ID!"
                    ],
                    "mergedId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "incrementOCount": [
                3,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "decrementOCount": [
                3,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "updateObservationStatus": [
                33,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "status": [
                        32,
                        "ObservationStatus!"
                    ]
                }
            ],
            "setPerformerFavorite": [
                34,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "favorite": [
                        4,
                        "Boolean!"
                    ]
                }
            ],
            "enrichPerformer": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "togglePlugin": [
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
                    "pluginName": [
                        1,
                        "String!"
                    ],
                    "values": [
                        39,
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
            "recordRecommendationChoice": [
                4,
                {
                    "chosenMediaId": [
                        2,
                        "ID!"
                    ],
                    "rejectedMediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "recordImpressions": [
                3,
                {
                    "impressions": [
                        41,
                        "[ImpressionInput!]!"
                    ]
                }
            ],
            "createSavedFilter": [
                50,
                {
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        48,
                        "SceneFilterInput!"
                    ]
                }
            ],
            "deleteSavedFilter": [
                4,
                {
                    "filterId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "refreshSceneComments": [
                52,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "recordSceneHeat": [
                4,
                {
                    "input": [
                        54,
                        "RecordSceneHeatInput!"
                    ]
                }
            ],
            "setSceneThumbnail": [
                4,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ],
                    "atSeconds": [
                        5,
                        "Float!"
                    ],
                    "jpegBase64": [
                        1,
                        "String!"
                    ]
                }
            ],
            "createSceneMarker": [
                56,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ],
                    "seconds": [
                        5,
                        "Float!"
                    ],
                    "endSeconds": [
                        5
                    ],
                    "tagName": [
                        1
                    ],
                    "label": [
                        1
                    ]
                }
            ],
            "deleteSceneMarker": [
                4,
                {
                    "markerId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "resolvePluginResult": [
                2,
                {
                    "pluginName": [
                        1,
                        "String!"
                    ],
                    "url": [
                        1,
                        "String!"
                    ],
                    "posterUrl": [
                        1
                    ]
                }
            ],
            "attachAlikeSource": [
                68,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ],
                    "pluginName": [
                        1,
                        "String!"
                    ],
                    "url": [
                        1,
                        "String!"
                    ]
                }
            ],
            "ensureEntityScenes": [
                3,
                {
                    "entityId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "recordSearch": [
                4,
                {
                    "query": [
                        1,
                        "String!"
                    ]
                }
            ],
            "forgetSearch": [
                4,
                {
                    "query": [
                        1,
                        "String!"
                    ]
                }
            ],
            "updateSettings": [
                62,
                {
                    "input": [
                        63,
                        "UpdateSettingsInput!"
                    ]
                }
            ],
            "upsertSettingEntry": [
                66,
                {
                    "input": [
                        67,
                        "UpsertSettingEntryInput!"
                    ]
                }
            ],
            "deleteSettingEntry": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "createStream": [
                68,
                {
                    "input": [
                        69,
                        "CreateStreamInput!"
                    ]
                }
            ],
            "deleteStream": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "ensureSceneStreams": [
                68,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "setStudioTags": [
                70,
                {
                    "studioId": [
                        2,
                        "ID!"
                    ],
                    "tagIds": [
                        2,
                        "[ID!]!"
                    ]
                }
            ],
            "subscribeSearch": [
                75,
                {
                    "query": [
                        1,
                        "String!"
                    ],
                    "sources": [
                        1,
                        "[String!]"
                    ],
                    "intervalHours": [
                        3
                    ]
                }
            ],
            "subscribe": [
                75,
                {
                    "kind": [
                        73,
                        "SubscriptionKind!"
                    ],
                    "targetId": [
                        2,
                        "ID!"
                    ],
                    "intervalHours": [
                        3
                    ]
                }
            ],
            "updateSearchSubscription": [
                75,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "intervalHours": [
                        3
                    ],
                    "enabled": [
                        4
                    ]
                }
            ],
            "unsubscribeSearch": [
                4,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "runSearchSubscription": [
                75,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "markSearchSubscriptionSeen": [
                75,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "upsertUserRating": [
                77,
                {
                    "input": [
                        78,
                        "UpsertUserRatingInput!"
                    ]
                }
            ],
            "deleteUserRating": [
                4,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "upsertWatchHistory": [
                79,
                {
                    "input": [
                        80,
                        "UpsertWatchHistoryInput!"
                    ]
                }
            ],
            "deleteWatchHistory": [
                4,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "addToWatchlist": [
                81,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "removeFromWatchlist": [
                4,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "Subscription": {
            "mediaAdded": [
                13
            ],
            "jobUpdated": [
                30
            ],
            "streamsChanged": [
                68,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "relatedChanged": [
                51,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "__typename": [
                1
            ]
        },
        "VpnStatus": {
            "connected": [
                4
            ],
            "interface": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "StreamResult": {
            "url": [
                1
            ],
            "mimeType": [
                1
            ],
            "quality": [
                1
            ],
            "headers": [
                10
            ],
            "loadTimeMs": [
                5
            ],
            "speedBps": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "StreamHeader": {
            "name": [
                1
            ],
            "value": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "ProviderSpeed": {
            "url": [
                1
            ],
            "loadTimeMs": [
                5
            ],
            "speedBps": [
                5
            ],
            "error": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "ScrapePayload": {
            "success": [
                4
            ],
            "error": [
                1
            ],
            "mediaId": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "MediaEvent": {
            "id": [
                2
            ],
            "type": [
                1
            ],
            "title": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "BlocklistEntry": {
            "id": [
                2
            ],
            "kind": [
                1
            ],
            "targetId": [
                2
            ],
            "label": [
                1
            ],
            "createdAt": [
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
            "pluginName": [
                1
            ],
            "username": [
                1
            ],
            "displayName": [
                1
            ],
            "sourceUrl": [
                1
            ],
            "thumbnailUrl": [
                1
            ],
            "tags": [
                1
            ],
            "gender": [
                1
            ],
            "subject": [
                1
            ],
            "viewers": [
                3
            ],
            "online": [
                4
            ],
            "lastSeenOnlineAt": [
                1
            ],
            "followed": [
                4
            ],
            "autoRecord": [
                4
            ],
            "recordQuality": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "CamRoomPage": {
            "rooms": [
                15
            ],
            "totalCount": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "CamSource": {
            "name": [
                1
            ],
            "displayName": [
                1
            ],
            "iconUrl": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "CamStream": {
            "url": [
                1
            ],
            "mimeType": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "CamRecording": {
            "id": [
                2
            ],
            "camModelId": [
                2
            ],
            "displayName": [
                1
            ],
            "thumbnailUrl": [
                1
            ],
            "quality": [
                1
            ],
            "status": [
                1
            ],
            "stopReason": [
                1
            ],
            "startedAt": [
                1
            ],
            "endedAt": [
                1
            ],
            "durationSeconds": [
                5
            ],
            "sizeBytes": [
                5
            ],
            "sceneId": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Change": {
            "id": [
                2
            ],
            "canonical": [
                2
            ],
            "observation": [
                2
            ],
            "plugin": [
                1
            ],
            "key": [
                1
            ],
            "action": [
                1
            ],
            "value": [
                1
            ],
            "originalValue": [
                1
            ],
            "iso6391": [
                1
            ],
            "iso31661": [
                1
            ],
            "score": [
                5
            ],
            "mode": [
                1
            ],
            "changedBy": [
                1
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Collection": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "details": [
                1
            ],
            "coverPath": [
                1
            ],
            "itemCount": [
                3
            ],
            "origin": [
                1
            ],
            "builtIn": [
                4
            ],
            "sourceUrl": [
                1
            ],
            "externalId": [
                1
            ],
            "tags": [
                76
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
        "CollectionMember": {
            "mediaId": [
                2
            ],
            "mediaType": [
                1
            ],
            "position": [
                3
            ],
            "title": [
                1
            ],
            "posterPath": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FaceIdentification": {
            "candidates": [
                27
            ],
            "videos": [
                24
            ],
            "posts": [
                25
            ],
            "postPerformers": [
                26
            ],
            "sources": [
                1
            ],
            "failures": [
                1
            ],
            "pictureUrl": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FaceVideoMatch": {
            "title": [
                1
            ],
            "similarity": [
                5
            ],
            "url": [
                1
            ],
            "thumbnailUrl": [
                1
            ],
            "sources": [
                1
            ],
            "scenePlugin": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FacePostMatch": {
            "title": [
                1
            ],
            "url": [
                1
            ],
            "thumbnailUrl": [
                1
            ],
            "solved": [
                4
            ],
            "names": [
                1
            ],
            "sources": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "FacePostPerformer": {
            "name": [
                1
            ],
            "posts": [
                3
            ],
            "imageUrl": [
                1
            ],
            "postUrl": [
                1
            ],
            "performerId": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "FaceCandidate": {
            "name": [
                1
            ],
            "similarity": [
                5
            ],
            "imageUrl": [
                1
            ],
            "profileUrls": [
                1
            ],
            "sources": [
                1
            ],
            "performerId": [
                2
            ],
            "__typename": [
                1
            ]
        },
        "Gallery": {
            "id": [
                2
            ],
            "externalId": [
                1
            ],
            "sourceUrl": [
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
            "coverPath": [
                1
            ],
            "imageCount": [
                3
            ],
            "rating": [
                5
            ],
            "organized": [
                4
            ],
            "studio": [
                70
            ],
            "performers": [
                34
            ],
            "tags": [
                76
            ],
            "images": [
                29
            ],
            "tagMatch": [
                1
            ],
            "video": [
                51
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
        "Image": {
            "id": [
                2
            ],
            "filePath": [
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
            "sourceUrl": [
                1
            ],
            "externalId": [
                1
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "position": [
                3
            ],
            "rating": [
                5
            ],
            "organized": [
                4
            ],
            "galleryId": [
                2
            ],
            "studio": [
                70
            ],
            "performers": [
                34
            ],
            "tags": [
                76
            ],
            "tagMatch": [
                1
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
        "Job": {
            "id": [
                2
            ],
            "kind": [
                1
            ],
            "status": [
                1
            ],
            "title": [
                1
            ],
            "description": [
                1
            ],
            "pluginName": [
                1
            ],
            "target": [
                2
            ],
            "error": [
                1
            ],
            "attempts": [
                3
            ],
            "maxAttempts": [
                3
            ],
            "priority": [
                3
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
            "createdAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "downloadTitle": [
                1
            ],
            "downloadUrl": [
                1
            ],
            "downloadProgress": [
                5
            ],
            "downloadBytesReceived": [
                5
            ],
            "downloadBytesTotal": [
                5
            ],
            "scene": [
                51
            ],
            "__typename": [
                1
            ]
        },
        "MediaCard": {
            "mediaId": [
                2
            ],
            "mediaType": [
                1
            ],
            "title": [
                1
            ],
            "posterPath": [
                1
            ],
            "backdropPath": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "ObservationStatus": {},
        "Observation": {
            "id": [
                2
            ],
            "target": [
                2
            ],
            "plugin": [
                1
            ],
            "sourceUrl": [
                1
            ],
            "confidence": [
                5
            ],
            "status": [
                32
            ],
            "data": [
                1
            ],
            "job": [
                2
            ],
            "observedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Performer": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "details": [
                1
            ],
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
            "careerLength": [
                1
            ],
            "url": [
                1
            ],
            "twitter": [
                1
            ],
            "instagram": [
                1
            ],
            "imagePath": [
                1
            ],
            "favorite": [
                4
            ],
            "rating": [
                5
            ],
            "tags": [
                76
            ],
            "sceneCount": [
                3
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
        "Plugin": {
            "id": [
                2
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
            "installedAt": [
                1
            ],
            "updatedAt": [
                1
            ],
            "settings": [
                37
            ],
            "settingValues": [
                38
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
        "PluginCategory": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "url": [
                1
            ],
            "count": [
                3
            ],
            "poster": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "ImpressionInput": {
            "mediaId": [
                2
            ],
            "source": [
                1
            ],
            "surface": [
                1
            ],
            "position": [
                3
            ],
            "clicked": [
                4
            ],
            "query": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "TasteProfile": {
            "signalCount": [
                3
            ],
            "tags": [
                43
            ],
            "performers": [
                43
            ],
            "studios": [
                43
            ],
            "sites": [
                43
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
            "imagePath": [
                1
            ],
            "affinity": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "RecommendedScene": {
            "scene": [
                51
            ],
            "source": [
                1
            ],
            "reason": [
                45
            ],
            "score": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "RecommendationReason": {
            "kind": [
                1
            ],
            "text": [
                1
            ],
            "entityId": [
                2
            ],
            "entityName": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "RecommendationRow": {
            "key": [
                1
            ],
            "title": [
                1
            ],
            "reason": [
                45
            ],
            "items": [
                44
            ],
            "__typename": [
                1
            ]
        },
        "RecommendedCategory": {
            "tag": [
                76
            ],
            "scenes": [
                51
            ],
            "__typename": [
                1
            ]
        },
        "SceneFilterInput": {
            "search": [
                1
            ],
            "studioId": [
                2
            ],
            "performerId": [
                2
            ],
            "tagId": [
                2
            ],
            "minRating": [
                5
            ],
            "minDuration": [
                3
            ],
            "maxDuration": [
                3
            ],
            "dateFrom": [
                1
            ],
            "dateTo": [
                1
            ],
            "sort": [
                1
            ],
            "sources": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SceneFilter": {
            "search": [
                1
            ],
            "studioId": [
                2
            ],
            "performerId": [
                2
            ],
            "tagId": [
                2
            ],
            "minRating": [
                5
            ],
            "minDuration": [
                3
            ],
            "maxDuration": [
                3
            ],
            "dateFrom": [
                1
            ],
            "dateTo": [
                1
            ],
            "sort": [
                1
            ],
            "sources": [
                1
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
                49
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Scene": {
            "id": [
                2
            ],
            "externalId": [
                1
            ],
            "sourceUrl": [
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
            "durationSeconds": [
                3
            ],
            "rating": [
                5
            ],
            "viewCount": [
                3
            ],
            "organized": [
                4
            ],
            "posterPath": [
                1
            ],
            "previewVideo": [
                1
            ],
            "previewImages": [
                1
            ],
            "studio": [
                70
            ],
            "performers": [
                34
            ],
            "tags": [
                76
            ],
            "tagMatch": [
                1
            ],
            "relatedVia": [
                1
            ],
            "related": [
                51,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "comments": [
                52
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
        "SceneComment": {
            "id": [
                2
            ],
            "plugin": [
                1
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
                34
            ],
            "__typename": [
                1
            ]
        },
        "SceneHeat": {
            "buckets": [
                5
            ],
            "bestMomentSeconds": [
                5
            ],
            "thumbnailSeconds": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "RecordSceneHeatInput": {
            "sceneId": [
                2
            ],
            "durationSeconds": [
                5
            ],
            "spans": [
                55
            ],
            "scrubs": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "HeatSpanInput": {
            "fromSeconds": [
                5
            ],
            "toSeconds": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "SceneMarker": {
            "id": [
                2
            ],
            "tag": [
                76
            ],
            "label": [
                1
            ],
            "seconds": [
                5
            ],
            "endSeconds": [
                5
            ],
            "personal": [
                4
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "PluginSearchResult": {
            "externalId": [
                1
            ],
            "title": [
                1
            ],
            "mediaType": [
                1
            ],
            "plugin": [
                1
            ],
            "sourceUrl": [
                1
            ],
            "date": [
                1
            ],
            "posterUrl": [
                1
            ],
            "previewImages": [
                1
            ],
            "previewVideo": [
                1
            ],
            "durationSeconds": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "AlikeCandidate": {
            "externalId": [
                1
            ],
            "title": [
                1
            ],
            "plugin": [
                1
            ],
            "sourceUrl": [
                1
            ],
            "posterUrl": [
                1
            ],
            "date": [
                1
            ],
            "previewImages": [
                1
            ],
            "previewVideo": [
                1
            ],
            "durationSeconds": [
                3
            ],
            "matchScore": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "SearchSuggestionKind": {},
        "SearchSuggestion": {
            "kind": [
                59
            ],
            "text": [
                1
            ],
            "entityId": [
                1
            ],
            "imageUrl": [
                1
            ],
            "detail": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SourceSuggestion": {
            "kind": [
                59
            ],
            "text": [
                1
            ],
            "sources": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Settings": {
            "maxConcurrentJobs": [
                3
            ],
            "maxJobRetries": [
                3
            ],
            "downloadSpeedLimitKBps": [
                3
            ],
            "allowDownloadsWhileStreaming": [
                4
            ],
            "camRecordingMaxMinutes": [
                3
            ],
            "autoEnrichAfterScrape": [
                4
            ],
            "requireVpn": [
                4
            ],
            "kindLimits": [
                64
            ],
            "__typename": [
                1
            ]
        },
        "UpdateSettingsInput": {
            "maxConcurrentJobs": [
                3
            ],
            "maxJobRetries": [
                3
            ],
            "downloadSpeedLimitKBps": [
                3
            ],
            "allowDownloadsWhileStreaming": [
                4
            ],
            "camRecordingMaxMinutes": [
                3
            ],
            "autoEnrichAfterScrape": [
                4
            ],
            "requireVpn": [
                4
            ],
            "kindLimits": [
                65
            ],
            "__typename": [
                1
            ]
        },
        "KindRateLimit": {
            "kind": [
                1
            ],
            "maxConcurrent": [
                3
            ],
            "retryInitialMs": [
                3
            ],
            "retryMultiplier": [
                5
            ],
            "retryMaxMs": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "KindRateLimitInput": {
            "kind": [
                1
            ],
            "maxConcurrent": [
                3
            ],
            "retryInitialMs": [
                3
            ],
            "retryMultiplier": [
                5
            ],
            "retryMaxMs": [
                3
            ],
            "__typename": [
                1
            ]
        },
        "SettingEntry": {
            "id": [
                2
            ],
            "scope": [
                1
            ],
            "plugin": [
                2
            ],
            "key": [
                1
            ],
            "value": [
                1
            ],
            "description": [
                1
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "UpsertSettingEntryInput": {
            "scope": [
                1
            ],
            "plugin": [
                2
            ],
            "key": [
                1
            ],
            "value": [
                1
            ],
            "description": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Stream": {
            "id": [
                2
            ],
            "mediaId": [
                2
            ],
            "url": [
                1
            ],
            "kind": [
                1
            ],
            "label": [
                1
            ],
            "provider": [
                1
            ],
            "resolution": [
                1
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "language": [
                1
            ],
            "format": [
                1
            ],
            "mimeType": [
                1
            ],
            "expectedSpeedBps": [
                5
            ],
            "fileSizeBytes": [
                5
            ],
            "verified": [
                4
            ],
            "pluginName": [
                1
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
        "CreateStreamInput": {
            "mediaId": [
                2
            ],
            "url": [
                1
            ],
            "kind": [
                1
            ],
            "label": [
                1
            ],
            "provider": [
                1
            ],
            "resolution": [
                1
            ],
            "width": [
                3
            ],
            "height": [
                3
            ],
            "language": [
                1
            ],
            "format": [
                1
            ],
            "mimeType": [
                1
            ],
            "expectedSpeedBps": [
                5
            ],
            "fileSizeBytes": [
                5
            ],
            "verified": [
                4
            ],
            "pluginName": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "Studio": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "aliases": [
                1
            ],
            "url": [
                1
            ],
            "parent": [
                70
            ],
            "imagePath": [
                1
            ],
            "details": [
                1
            ],
            "tags": [
                76
            ],
            "sceneCount": [
                3
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
        "SubscriptionFeedFilter": {
            "subscriptionId": [
                2
            ],
            "kinds": [
                73
            ],
            "newOnly": [
                4
            ],
            "unwatchedOnly": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "SubscriptionFeedItem": {
            "scene": [
                51
            ],
            "foundAt": [
                1
            ],
            "subscriptionId": [
                2
            ],
            "isNew": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "SubscriptionKind": {},
        "SubscriptionTarget": {
            "id": [
                2
            ],
            "name": [
                1
            ],
            "imageUrl": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "SearchSubscription": {
            "id": [
                2
            ],
            "kind": [
                73
            ],
            "query": [
                1
            ],
            "target": [
                74
            ],
            "sources": [
                1
            ],
            "intervalHours": [
                3
            ],
            "enabled": [
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
            "newCount": [
                3
            ],
            "totalCount": [
                3
            ],
            "createdAt": [
                1
            ],
            "scenes": [
                51,
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
            "description": [
                1
            ],
            "category": [
                1
            ],
            "sceneCount": [
                3
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
        "UserRating": {
            "id": [
                2
            ],
            "media": [
                2
            ],
            "rating": [
                5
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
        "UpsertUserRatingInput": {
            "media": [
                2
            ],
            "rating": [
                5
            ],
            "__typename": [
                1
            ]
        },
        "WatchHistory": {
            "id": [
                2
            ],
            "media": [
                2
            ],
            "scene": [
                51
            ],
            "startedAt": [
                1
            ],
            "finishedAt": [
                1
            ],
            "progressSeconds": [
                3
            ],
            "maxProgressSeconds": [
                3
            ],
            "durationSeconds": [
                3
            ],
            "completed": [
                4
            ],
            "updatedAt": [
                1
            ],
            "__typename": [
                1
            ]
        },
        "UpsertWatchHistoryInput": {
            "media": [
                2
            ],
            "progressSeconds": [
                3
            ],
            "durationSeconds": [
                3
            ],
            "completed": [
                4
            ],
            "__typename": [
                1
            ]
        },
        "WatchlistItem": {
            "id": [
                2
            ],
            "media": [
                2
            ],
            "createdAt": [
                1
            ],
            "__typename": [
                1
            ]
        }
    }
}