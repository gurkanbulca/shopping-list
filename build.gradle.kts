plugins {
    alias(libs.plugins.androidApplication) apply false
    alias(libs.plugins.androidLibrary) apply false
    alias(libs.plugins.kotlinAndroid) apply false
    alias(libs.plugins.kotlinMultiplatform) apply false
    alias(libs.plugins.sqldelight) apply false
    alias(libs.plugins.kotlinSerialization) apply false
}

allprojects {
    group = "shopping"
    version = "1.0.0-SNAPSHOT"
}

tasks.register("clean", Delete::class) {
    delete(rootProject.layout.buildDirectory)
}
