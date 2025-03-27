#version 330

uniform sampler2DArray texU;
uniform sampler2DArray texV;

in vec2 tex_coord;

layout(location = 0) out vec2 outUV;
layout(location = 1) out float outV;

void main()
{
    vec4 in_colorU = texture(texU, vec3(tex_coord, 0));
    vec4 in_colorV = texture(texV, vec3(tex_coord, 0));

    outUV = vec2(in_colorU.r, in_colorV.r);
    outV = in_colorV.r;
}