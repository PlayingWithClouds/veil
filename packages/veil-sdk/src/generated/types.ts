export default {
    "scalars": [
        1,
        2,
        3,
        4,
        5,
        22,
        48,
        61
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
            "changes": [
                15,
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
                16,
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
                16,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "collectionMembers": [
                17,
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
                18,
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
                18,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "images": [
                19,
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
                19,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "jobs": [
                20
            ],
            "job": [
                20,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "downloadedScenes": [
                41
            ],
            "mediaCards": [
                21,
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
                23,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "observations": [
                23,
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
                24,
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
                24,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "plugins": [
                26
            ],
            "plugin": [
                26,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "pluginPackages": [
                25,
                {
                    "query": [
                        1
                    ]
                }
            ],
            "pluginCategories": [
                30,
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
                34,
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
                32,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "recommendedRows": [
                36,
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
                37,
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
                41
            ],
            "savedFilters": [
                40
            ],
            "scenes": [
                41,
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
                41,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "randomScenes": [
                41,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "recommendedFeed": [
                41,
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
                42,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "sceneMarkers": [
                45,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "pluginSearch": [
                46,
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
            "recommendedBrowse": [
                46,
                {
                    "limit": [
                        3
                    ]
                }
            ],
            "findAlikeSources": [
                47,
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
                49,
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
                50
            ],
            "settingEntries": [
                54,
                {
                    "scope": [
                        1
                    ]
                }
            ],
            "mediaStreams": [
                56,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "studios": [
                58,
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
                58,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "searchSubscriptions": [
                63
            ],
            "searchSubscription": [
                63,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "subscriptionForTarget": [
                63,
                {
                    "targetId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "subscriptionFeed": [
                60,
                {
                    "filter": [
                        59
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
                64,
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
                64,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "userRating": [
                65,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "userRatings": [
                65,
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
                67,
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
                67,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "watchlist": [
                69
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
            "createCollection": [
                16,
                {
                    "name": [
                        1,
                        "String!"
                    ]
                }
            ],
            "renameCollection": [
                16,
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
                16,
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
            "ensureGalleryImages": [
                18,
                {
                    "galleryId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "setImageTags": [
                19,
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
                23,
                {
                    "id": [
                        2,
                        "ID!"
                    ],
                    "status": [
                        22,
                        "ObservationStatus!"
                    ]
                }
            ],
            "setPerformerFavorite": [
                24,
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
                        29,
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
                        31,
                        "[ImpressionInput!]!"
                    ]
                }
            ],
            "createSavedFilter": [
                40,
                {
                    "name": [
                        1,
                        "String!"
                    ],
                    "filter": [
                        38,
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
            "recordSceneHeat": [
                4,
                {
                    "input": [
                        43,
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
                45,
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
                56,
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
                50,
                {
                    "input": [
                        51,
                        "UpdateSettingsInput!"
                    ]
                }
            ],
            "upsertSettingEntry": [
                54,
                {
                    "input": [
                        55,
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
                56,
                {
                    "input": [
                        57,
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
                56,
                {
                    "sceneId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "setStudioTags": [
                58,
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
                63,
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
                63,
                {
                    "kind": [
                        61,
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
                63,
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
                63,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "markSearchSubscriptionSeen": [
                63,
                {
                    "id": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "upsertUserRating": [
                65,
                {
                    "input": [
                        66,
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
                67,
                {
                    "input": [
                        68,
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
                69,
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
                20
            ],
            "streamsChanged": [
                56,
                {
                    "mediaId": [
                        2,
                        "ID!"
                    ]
                }
            ],
            "relatedChanged": [
                41,
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
            "sourceUrl": [
                1
            ],
            "externalId": [
                1
            ],
            "tags": [
                64
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
                58
            ],
            "performers": [
                24
            ],
            "tags": [
                64
            ],
            "images": [
                19
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
                58
            ],
            "performers": [
                24
            ],
            "tags": [
                64
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
                41
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
                22
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
                64
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
                27
            ],
            "settingValues": [
                28
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
            "__typename": [
                1
            ]
        },
        "TasteProfile": {
            "signalCount": [
                3
            ],
            "tags": [
                33
            ],
            "performers": [
                33
            ],
            "studios": [
                33
            ],
            "sites": [
                33
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
                41
            ],
            "source": [
                1
            ],
            "reason": [
                35
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
                35
            ],
            "items": [
                34
            ],
            "__typename": [
                1
            ]
        },
        "RecommendedCategory": {
            "tag": [
                64
            ],
            "scenes": [
                41
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
                39
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
                58
            ],
            "performers": [
                24
            ],
            "tags": [
                64
            ],
            "tagMatch": [
                1
            ],
            "related": [
                41,
                {
                    "limit": [
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
                44
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
                64
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
                48
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
            "autoEnrichAfterScrape": [
                4
            ],
            "requireVpn": [
                4
            ],
            "kindLimits": [
                52
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
            "autoEnrichAfterScrape": [
                4
            ],
            "requireVpn": [
                4
            ],
            "kindLimits": [
                53
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
                58
            ],
            "imagePath": [
                1
            ],
            "details": [
                1
            ],
            "tags": [
                64
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
                61
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
                41
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
                61
            ],
            "query": [
                1
            ],
            "target": [
                62
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
                41,
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
                41
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