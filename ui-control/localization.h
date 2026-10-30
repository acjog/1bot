#pragma once
#include <QtCore>

class Localization {
public:
    Localization();
    bool setLanguage(const QString &preference);
    static QString resolveLanguage(const QString &preference);
private:
    QTranslator applicationTranslator;
    QTranslator qtTranslator;
};
