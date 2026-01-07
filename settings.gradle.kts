pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "ShoppingList"

// Shared KMP modules
include(":shared:domain")
include(":shared:data")
include(":shared:db")
include(":shared:sync")
include(":shared:platform")

// Native app modules
include(":androidApp")
