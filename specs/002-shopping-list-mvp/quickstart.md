# Quickstart: Mobile Offline MVP

**Status**: Placeholder  
**Feature**: Mobile Offline MVP (Android + iOS)

## Overview

This document provides a quick start guide for testing and validating the mobile app implementation.

## Prerequisites

- Android Studio (for Android development)
- Xcode 15+ (for iOS development)
- JDK 17+
- Backend server running (see main README)

## Quick Start

### 1. Clone and Setup

```bash
# Clone repository (if not already done)
git clone <repository-url>
cd shopping-list

# Ensure backend is running
docker-compose up -d
```

### 2. Build Android App

```bash
./gradlew :androidApp:assembleDebug
./gradlew :androidApp:installDebug
```

### 3. Build iOS App

```bash
open iosApp/ShoppingListApp.xcodeproj
# Build and run from Xcode
```

## Smoke Test Checklist

### User Story 1: Sign in and Browse

- [ ] Fresh install: App shows login screen
- [ ] Login: Enter credentials, submit → Groups screen appears
- [ ] Groups: Select a group → Lists screen appears
- [ ] Lists: Select a list → Items screen appears
- [ ] Items: Items display in correct sort order (unpurchased first, then by priority)
- [ ] Restart: Close app, reopen → User remains logged in

### User Story 2: Offline Add/Toggle

- [ ] Disable network (airplane mode)
- [ ] Add item: Item appears immediately with pending indicator
- [ ] Toggle purchased: State changes immediately with pending indicator
- [ ] Enable network: Pending indicators clear, server confirms changes

### User Story 3: Reorder and Conflict

- [ ] Reorder items: New order persists across navigation
- [ ] Offline reorder: Queued and syncs when online
- [ ] Conflict scenario: App refreshes and allows retry

## Configuration

### API Endpoint

Configure the backend endpoint in:
- Android: `androidApp/src/main/java/shopping/android/grpc/GrpcConfig.kt`
- iOS: `iosApp/ShoppingListApp/GrpcConfig.swift`

### Debug Logging

Enable verbose logging for debugging sync issues:
- Android: Set `SHOPPING_DEBUG=true` in build config
- iOS: Set `SHOPPING_DEBUG` environment variable

## Troubleshooting

### Common Issues

1. **Build fails with Gradle errors**
   - Ensure JDK 17+ is installed and `JAVA_HOME` is set
   - Run `./gradlew clean` and retry

2. **iOS build fails**
   - Ensure Xcode command line tools are installed
   - Check that iOS deployment target matches device

3. **Network errors on Android emulator**
   - Use `10.0.2.2` instead of `localhost` for backend
   - Check emulator network settings

## Notes

- TODO: Add detailed testing scenarios
- TODO: Add performance benchmarks
- TODO: Add error scenario testing
