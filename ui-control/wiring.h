#pragma once
#include <QtCore>
int runWiring();
QByteArray protectSecret(const QByteArray &value);
QByteArray unprotectSecret(const QByteArray &value);
