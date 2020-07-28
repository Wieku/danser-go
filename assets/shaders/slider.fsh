#version 330

//#define borderEnd 99f/512f //79
#define borderStart 0.06640625f // 34/512
#define baseBorderWidth 0.126953125f // 65/512
#define blend 0.01f

#define maxBorderWidth 1.0f - borderStart

#define slope (maxBorderWidth - baseBorderWidth) / 9

uniform vec4 col_border;
uniform vec4 col_border1;

in float distance_inv;
out vec4 color;

const float borderWidth = 1.0f;

void main()
{
    vec4 borderColorOuter = col_border1;
    vec4 borderColorInner = col_border;
    vec4 outerShadow = vec4(vec3(0.0), 0.5 * distance_inv / borderStart * borderColorInner.a);
    vec4 bodyColorOuter = vec4(vec3(0.05), borderColorInner.a);
    vec4 bodyColorInner = vec4(vec3(0.2), borderColorInner.a);

    float borderWidthScaled = borderWidth < 0 ? borderWidth * baseBorderWidth : (borderWidth - 1.0f) * slope + baseBorderWidth;
    float borderMid = borderStart + borderWidthScaled / 2;
    float borderEnd = borderStart + borderWidthScaled;

    vec4 borderColorMix = mix(borderColorOuter, borderColorInner, smoothstep(borderMid - borderWidthScaled/4, borderMid + borderWidthScaled/4, distance_inv));
    vec4 bodyColorMix = mix(bodyColorOuter, bodyColorInner, (distance_inv - borderEnd) / (1f - borderEnd));

    if (borderWidth < 0.01) {
        borderColorMix = outerShadow;
    }

    if (borderWidth > 9.99f) {
        bodyColorMix = borderColorMix;
    }


    if (distance_inv <= borderStart - blend) {
        color = outerShadow;
    }

    if (distance_inv > borderStart-blend && distance_inv < borderStart+blend) {
        color = mix(outerShadow, borderColorMix, (distance_inv - (borderStart - blend)) / (2 * blend));
    }

    if (distance_inv > borderStart+blend && distance_inv <= borderEnd-blend) {
        color = borderColorMix;
    }

    if (distance_inv > borderEnd-blend && distance_inv < borderEnd+blend) {
        color = mix(borderColorMix, bodyColorMix, (distance_inv - (borderEnd - blend)) / (2 * blend));
    }

    if (distance_inv > borderEnd + blend) {
        color = bodyColorMix;
    }
}